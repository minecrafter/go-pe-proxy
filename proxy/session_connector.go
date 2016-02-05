package proxy

import (
  "bytes"
  "encoding/hex"
  "fmt"
  "math/rand"
  "net"
  "log"
  "sync"
  "time"
  "../util"
  "../packets/raknet"
  "../packets/mcpe"
)

type SessionConnector struct {
  session *Session
  server *Server

  conn *net.UDPConn

  // INTERNAL
  reliabilityNumber util.AtomicInteger
  datagramSequenceNumber util.AtomicInteger
  splitPackets raknet.SplitPacketHandler
  firstServer bool
  // INTERNAL: timer used for periodic tick task
  fastTimer *time.Ticker
  slowTimer *time.Ticker
  // INTERNAL: used to communicate acks
  ackQueue []int32
  ackQueueLock sync.Mutex
  mtu int16
  guid int64
  state sessionConnectorState

  // INTERNAL: Used to communicate player packets to the backend
  packetQueue chan []byte
  // INTERNAL: Used only when the connection is to be closed
  closeChan chan struct{}
  abandoned bool
}

func NewSessionConnector(session *Session, server *Server) (this *SessionConnector) {
  this = new(SessionConnector)
  this.session = session
  this.server = server
  this.guid = rand.Int63()
  this.splitPackets = raknet.NewSplitPacketHandler()
  this.fastTimer = time.NewTicker(50 * time.Millisecond) // MiNET uses this
  this.slowTimer = time.NewTicker(5 * time.Second)
  this.packetQueue = make(chan []byte, 300) // _more_ than enough!
  this.closeChan = make(chan struct{})

  if session.serverConnection == nil {
    this.firstServer = true
  }

  this.state = C_STATE_UNCONNECTED

  // Try to use the same MTU as the client
  this.mtu = session.mtu

  return
}

func (this *SessionConnector) IsAlive() bool {
  return !this.abandoned
}

func (this *SessionConnector) GetEndpointString() string {
  return this.server.Name
}

func (this *SessionConnector) Close() (err error) {
  this.fastTimer.Stop()
  this.slowTimer.Stop()
  this.abandoned = true

  // Send a disconnect packet
  err = this.SendPacket(raknet.RakNetDisconnectNotification{})
  if err != nil {
    return
  }

  // Kill the goroutine processing connections
  this.closeChan <- struct{}{}

  // Close the UDP connection
  err = this.conn.Close()
  return
}

func (this *SessionConnector) Process() {
  for {
    select {
    case <-this.closeChan:
      close(this.packetQueue)
      close(this.closeChan)
      return
    case <-this.slowTimer.C:
      pkt := raknet.NewConnectedPingWithCurrentTime()
      if err := this.SendPackage(pkt); err != nil {
        // TODO: Handle gracefully
      }
    case pkt := <-this.packetQueue:
      // Reframe all received packets.
      all, err := this.handleDatagram(pkt)
      if err != nil {
        log.Printf("Unable to reassemble packet for %s: %s", this.server.Address.String(), err.Error())
        break
      }
      // Nothing? Then return, we'll have something soon.
      if all == nil {
        break
      }

      for _, p := range *all {
        repackaged := raknet.GenericRakNetPackage{
          PacketId: p[0],
          Payload: p[1:],
        }
        if err = this.session.SendPackage(repackaged); err != nil {
          log.Printf("Unable to send packet to %s: %s", this.server.Address.String(), err.Error())
          break
        }
      }
    case <-this.fastTimer.C:
      this.splitPackets.GarbageCollect()

      // Drain acks
      this.ackQueueLock.Lock()
      var toAck []int
      for item := range this.ackQueue {
        toAck = append(toAck, int(item))
      }
      this.ackQueue = make([]int32, 0)
      this.ackQueueLock.Unlock()

      if len(toAck) > 0 {
        ackPkt := new(raknet.RakNetAck)
        ackPkt.Acknowledged = raknet.SliceAck(toAck)
        if err := this.SendPacket(ackPkt); err != nil {
          log.Printf("Unable to send acks for %s: %s", this.server.Address.String(), err.Error())
        }
      }
    }
  }
}

func (this *SessionConnector) Connect() (err error) {
  c, err := net.DialUDP("udp4", nil, this.server.Address)

  if err != nil {
    return
  }

  this.conn = c
  this.state = C_STATE_IDENTIFY
  go this.connectionListener()

  // Send the first handshake
  first := raknet.RakNetOpenConnectionRequest1{7, this.mtu - 32} // ????
  err = this.SendPacket(first)
  return
}

func (this *SessionConnector) SendPacket(pkt raknet.EncodablePacket) error {
  log.Println("Sending to backend:", pkt)
  return raknet.WriteUDPPreConnected(this.conn, pkt)
}

func (this *SessionConnector) SendPackage(pkg raknet.EncodablePacket) error {
  encapsulated, err := raknet.CreateDatagrams(&this.reliabilityNumber,
    &this.datagramSequenceNumber, pkg, this.mtu)
  if err != nil {
    return err
  }

  for _, item := range *encapsulated {
    err = this.SendPacket(item)
    if err != nil {
      return err
    }
  }

  return nil
}

func (this *SessionConnector) connectionListener() {
  for {
    buf := make([]byte, this.mtu)
    read, _, err := this.conn.ReadFromUDP(buf)

    if err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }

    pktBytes := buf[0:read]

    this.dispatchData(pktBytes)
  }
}

func (this *SessionConnector) dispatchData(pktBytes []byte) {
  //log.Printf("Backend DISPATCHED: %s", hex.EncodeToString(pktBytes))
  if this.state == C_STATE_IDENTIFY {
    this.handleIdentify(pktBytes)
  } else if this.state == C_STATE_CONNECTED {
    this.handleConnected(pktBytes)
  }
}

func (this *SessionConnector) handleMcpeBatch(pktData []byte) {
  //log.Println("Backend packet data:", hex.EncodeToString(pktData))

  pkt := new(mcpe.MCPEBatch)
  err := pkt.Decode(bytes.NewReader(pktData))
  if err != nil {
    log.Printf("Error whilst handling backend message: %s", err)
    return
  }

  for _, item := range pkt.Payload {
    log.Printf("Handling decompressed packet: %s ...", hex.EncodeToString(item[0:15]))
    this.dispatchData(item)
  }
}

func (this *SessionConnector) handleDatagram(pktBytes []byte) (*[][]byte, error) {
  pkt := new(raknet.RakNetDatagram)
  err := pkt.Decode(pktBytes)
  if err != nil {
    return nil, err
  }

  defer func() {
    this.ackQueueLock.Lock()
    this.ackQueue = append(this.ackQueue, pkt.DatagramSequenceNumber)
    this.ackQueueLock.Unlock()
  }()

  var all [][]byte

  for _, item := range pkt.Payload {
    if item.PartCount > 1 {
      if allPackets := this.splitPackets.AcceptSplitPacket(*item); allPackets != nil {
        // We have all the packets. Reconstruct them and handle it.
        // TODO: Handle out-of-order packets.
        var b bytes.Buffer
        for _, p := range *allPackets {
          b.Write(p.Payload)
        }

        all = append(all, b.Bytes())
      }
    } else {
      all = append(all, item.Payload)
    }
  }

  return &all, nil
}

func (this *SessionConnector) handleDatagramIdentify(pktBytes []byte) {
  all, err := this.handleDatagram(pktBytes)
  if err != nil {
    log.Println("Unable to handle datagram from backend:", err)
    return
  }
  if len(*all) > 0 {
    for _, item := range *all {
      this.handleIdentify(item)
    }
  }
}

func (this *SessionConnector) handleIdentify(pktBytes []byte) (err error) {
  pktData := pktBytes[1:]

  switch pktBytes[0] {
  case raknet.ID_DATA_4:
    this.handleDatagramIdentify(pktBytes)
  case raknet.ID_DATA_C:
    this.handleDatagramIdentify(pktBytes)
  case raknet.ID_OPEN_CONNECTION_REPLY_1:
    pkt := new(raknet.RakNetOpenConnectionReply1)
    err = pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      return
    }

    // Use server-provided MTU
    this.mtu = pkt.MTU

    // Send the second request packet.
    reply := raknet.RakNetOpenConnectionRequest2{
      GUID: this.guid,
      ClientEndpoint: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 19132},
      MTU: this.mtu,
    }

    if err = this.SendPacket(reply); err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }
  case raknet.ID_OPEN_CONNECTION_REPLY_2:
    // Ignore. Move on to requesting a connection.
    reply := raknet.RakNetConnectionRequest{
      GUID: this.guid,
      Timestamp: raknet.GetTimeMilliseconds(),
      Security: 0,
    }

    if err = this.SendPackage(reply); err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }
  case raknet.ID_CONNECTION_REQUEST_ACCEPTED:
    // Ignore this too. Send a login packet.
    r1 := raknet.RakNetNewIncomingConnection{
      Cookie: rand.Int31(),
      Secure: 0,
      Port: 19132,
      Session1: 42,
      Session2: 42,
    }
    if err = this.SendPackage(r1); err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }

    batchPkt := new(mcpe.MCPEBatch)
    if err = batchPkt.AddPacket(this.session.loginPkt); err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }

    if err = this.SendPackage(batchPkt); err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }
  case mcpe.ID_MCPE_DISCONNECT:
    // In spite of all attempts, this connection has failed.
    // Forward this message to the client, and close this connection.
    pkt := new(mcpe.MCPEDisconnect)
    err = pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      log.Println("Encountered an error while handling backend connection:", err)
      // TODO: Graceful handling of this situation.
      return
    }

    // TODO: Would forwarding be more approriate?
    log.Printf("%s disconnected from %s: %s", this.session.endpoint.String(), this.server.Name, pkt.Message)
    this.session.AbandonWithReason(fmt.Sprintf("Disconnected from %s: %s", this.server.Name, pkt.Message))
  case mcpe.ID_MCPE_START_GAME:
    pkt := new(mcpe.MCPEStartGame)
    err = pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      return
    }

    // If this is our first server, we'll simply forward this packet on.
    // If it isn't, we'll send a respawn packet instead.
    // TODO: Implement this properly.
    go this.Process()
    this.state = C_STATE_CONNECTED
    this.session.state = STATE_CONNECTED
    err = this.session.SendDirect(pktBytes)
    if err != nil {
      return
    }
  }

  return nil
}

func (this *SessionConnector) handleConnected(pktBytes []byte) (err error) {
  // In case of ACK or NAK, don't send them to the client.
  if pktBytes[0] == raknet.ID_ACK || pktBytes[0] == raknet.ID_NAK {
    return
  }

  // Generally, we won't meddle with connected player's packets, except to
  // rewrite entity IDs.

  // NOTE: After initial authentication, we'll be sent datagrams. MiNET batches
  // these into a single MCPEBatch datagram, requiring us to decompress the packet,
  // read the data (if needed), and then reconstruct the final data. How we will deal
  // with other servers (i.e. PocketMine-MP) is unknown.

  // TODO: Implement entity ID rewriting. For now, all we can do is repackage the
  // datagram ourselves.
  if pktBytes[0] == raknet.ID_DATA_4 || pktBytes[0] == raknet.ID_DATA_C {
    payload, err := this.handleDatagram(pktBytes)
    if err != nil {
      return err
    }

    // Nothing? Then return, we'll have something soon.
    if payload == nil {
      return nil
    }

    for _, p := range *payload {
      repackaged := raknet.GenericRakNetPackage{
        PacketId: p[0],
        Payload: p[1:],
      }
      if err = this.session.SendPackage(repackaged); err != nil {
        return err
      }
    }
  }

  return
}
