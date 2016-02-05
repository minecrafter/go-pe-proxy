package proxy

import (
  "bytes"
  "encoding/hex"
  "fmt"
  "log"
  "net"
  "sync"
  "time"
  "github.com/pborman/uuid"
  "../packets/mcpe"
  "../packets/raknet"
  "../util"
)

type Session struct {
  // Session information
  username *string
  uuid *uuid.UUID
  endpoint *net.UDPAddr
  proxy *Proxy
  mtu int16
  serverConnection *SessionConnector

  // INTERNAL: channel used to send packets for processing
  processQueue chan []byte
  // INTERNAL: timer used for periodic tick task
  timer *time.Ticker
  // INTERNAL: channel used to stop the player goroutine
  poison chan struct{}
  // INTERNAL: used to communicate acks
  datagramHelper *raknet.DatagramHelper
  ackQueue []int32
  ackQueueLock sync.Mutex
  // INTERNAL
  state sessionState
  // INTERNAL: marks a session as abandoned
  abandoned bool
  // INTERNAL: last ping message received time
  lastPing time.Time
  // INTERNAL
  reliabilityNumber util.AtomicInteger
  datagramSequenceNumber util.AtomicInteger
  // INTERNAL
  splitPackets raknet.SplitPacketHandler
  // INTERNAL: storing MCPELogin packet here so we can forward it
  loginPkt *mcpe.MCPELogin
  // INTERNAL: stores last known dimension
  dimension byte
}

func NewSession(proxy *Proxy, mtu int16, endpoint *net.UDPAddr) (this *Session) {
  this = new(Session)
  this.proxy = proxy
  this.mtu = mtu
  this.endpoint = endpoint
  this.state = STATE_IDENTIFY
  this.splitPackets = raknet.NewSplitPacketHandler()
  this.abandoned = false
  this.lastPing = time.Now() // otherwise the client gets d/c'ed

  this.processQueue = make(chan []byte, 300) // 300 should be more than enough
                                             // unless the client is very laggy
  this.poison = make(chan struct{}, 1)
  this.timer = time.NewTicker(50 * time.Millisecond) // MiNET uses this
  this.ackQueue = make([]int32, 0) // this should be enough
  this.datagramHelper = raknet.NewDatagramHelper(this)

  return
}

func (this *Session) IsAlive() bool {
  return !this.abandoned
}

func (this *Session) GetEndpointString() string {
  return this.endpoint.String()
}

func (this *Session) handleSession() {
  for {
    select {
    case pktBytes := <-this.processQueue:
      this.lastPing = time.Now()

      // Generic: Always respond to these.
      pktData := pktBytes[1:]
      switch pktBytes[0] {
        case raknet.ID_CONNECTED_PING:
          pkt := new(raknet.RakNetConnectedPing)
          err := pkt.Decode(bytes.NewReader(pktData))
          if err != nil {
            log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
            break
          }

          log.Printf("Handling a connected ping packet from %s.", this.endpoint.String())
          reply := raknet.RakNetConnectedPong{
            Timestamp1: pkt.Timestamp,
            Timestamp2: raknet.GetTimeMilliseconds(),
          }
          if err = this.SendPacket(reply); err != nil {
            log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
            break
          }
        case raknet.ID_DISCONNECT_NOTIFICATION:
          // Player wants to disconnect. Signal an abandoned connection.
          // This very same goroutine will pick it up and actually abandon this
          // connection.
          this.Abandon()
        case raknet.ID_ACK:
          pkt := new(raknet.RakNetAck)
          err := pkt.Decode(bytes.NewReader(pktData))
          if err != nil {
            log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
            break
          }

          this.datagramHelper.HandleAck(pkt)
        case raknet.ID_NAK:
          pkt := new(raknet.RakNetNak)
          err := pkt.Decode(bytes.NewReader(pktData))
          if err != nil {
            log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
            break
          }

          this.datagramHelper.HandleNak(pkt)
        default:
          // Otherwise, we have special-cased packets. It's better to handle these in
          // a separate internal function.
          this.dispatchData(pktBytes)
      }
    // TODO: This will cause connections to be killed.
    case <-this.timer.C:
      this.splitPackets.GarbageCollect()
      this.datagramHelper.TryResendPackets()

      if this.lastPing.Add(10 * time.Second).Before(time.Now()) {
        // disconnect
        log.Printf("Ping timeout for %s", this.endpoint.String())
        this.AbandonWithReason("Ping timeout")
      } else {
        // Drain acks
        /*this.ackQueueLock.Lock()
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
            log.Printf("Unable to send acks for %s: %s", this.endpoint.String(), err.Error())
          }
        }*/
      }
    case <-this.poison:
      // Commit suicide.
      this.abandoned = true
      close(this.processQueue)
      close(this.poison)
      this.timer.Stop()
      return
    }
  }
}

func (this *Session) SendPackage(pkg raknet.EncodablePacket) error {
  encapsulated, err := raknet.CreateDatagrams(&this.reliabilityNumber,
    &this.datagramSequenceNumber, pkg, this.mtu)
  if err != nil {
    return err
  }

  for _, item := range *encapsulated {
    this.datagramHelper.RegisterDatagram(item)
    err = this.SendPacket(item)
    if err != nil {
      return err
    }
  }

  return nil
}

func (this *Session) SendDirect(pkt []byte) (err error) {
  _, err = this.proxy.conn.WriteToUDP(pkt, this.endpoint)
  return
}

func (this *Session) SendPacket(pkt raknet.EncodablePacket) (error) {
  return raknet.WriteUDP(this.proxy.conn, this.endpoint, pkt)
}

func (this *Session) Connect(server *Server) {
  connector := NewSessionConnector(this, server)
  firstServer := true

  if this.serverConnection != nil {
    firstServer = false
  }

  err := connector.Connect()
  if err != nil {
    msg := fmt.Sprintf("Unable to connect to %s: %s", server.Name, err.Error())
    if firstServer {
      this.AbandonWithReason(msg)
    } else {
      chat := mcpe.MCPEText{mcpe.TEXT_TYPE_RAW, "", msg}
      if e := this.SendPackage(chat); e != nil {
        // TODO: What?
      }
    }
  }
}

func (this *Session) AbandonWithReason(reason string) bool {
  if this.abandoned {
    return false // already abandoned!
  }

  pkt := mcpe.MCPEDisconnect{reason}
  if err := this.SendPackage(pkt); err != nil {
    return false
  }

  return this.Abandon()
}

func (this *Session) Abandon() bool {
  if this.abandoned {
    return false // already abandoned!
  }

  // Unregister ourselves
  this.proxy.Registry.Unregister(this)

  // Cancel the player's goroutine task
  // This will also close the processQueue channel
  this.poison <- struct{}{}

  // If the player is connected, abandon their connection too.
  if this.serverConnection != nil {
    this.serverConnection.Close()
  }

  return true
}

func safeSlice(p []byte, end int) []byte {
  ln := len(p)
  if end >= ln {
    end = ln - 1
  }
  return p[0:end]
}

func (this *Session) dispatchData(pktBytes []byte) {
  log.Printf("DISPATCHED: %s", hex.EncodeToString(safeSlice(pktBytes, 32)))
  if this.state == STATE_IDENTIFY {
    this.handleIdentify(pktBytes)
  }
}

func (this *Session) handleMcpeBatch(pktData []byte) {
  log.Println("Our packet data:", hex.EncodeToString(pktData))

  pkt := new(mcpe.MCPEBatch)
  err := pkt.Decode(bytes.NewReader(pktData))
  if err != nil {
    log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
    return
  }

  for _, item := range pkt.Payload {
    log.Printf("Handling decompressed packet: %s ...", hex.EncodeToString(item[0:15]))
    this.dispatchData(item)
  }
}

func (this *Session) handleDatagram(pktBytes []byte) {
  pkt := new(raknet.RakNetDatagram)
  err := pkt.Decode(pktBytes)
  if err != nil {
    log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
    return
  }

  defer func() {
    this.ackQueueLock.Lock()
    this.ackQueue = append(this.ackQueue, pkt.DatagramSequenceNumber)
    this.ackQueueLock.Unlock()
  }()

  log.Println("Datagram: ", pkt)

  for _, item := range pkt.Payload {
    log.Println("Datagram data part: ", *item)
    if item.PartCount > 1 {
      if allPackets := this.splitPackets.AcceptSplitPacket(*item); allPackets != nil {
        // We have all the packets. Reconstruct them and handle it.
        // TODO: Handle out-of-order packets.
        var b bytes.Buffer
        for _, p := range *allPackets {
          b.Write(p.Payload)
        }

        f := b.Bytes()
        this.dispatchData(f)
      }
    } else {
      this.dispatchData(item.Payload)
    }
  }
}

func (this *Session) handleIdentify(pktBytes []byte) {
  pktData := pktBytes[1:]

  switch pktBytes[0] {
  case raknet.ID_DATA_4:
    this.handleDatagram(pktBytes)
  case raknet.ID_DATA_C:
    this.handleDatagram(pktBytes)
  case raknet.ID_CONNECTION_REQUEST:
    pkt := new(raknet.RakNetConnectionRequest)
    err := pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
      return
    }

    // Have to wrap this in a datagram.
    toEncapsulate := raknet.RakNetConnectionRequestAccepted{
      SystemAddress: net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 19312},
      IncomingTimestamp: pkt.Timestamp,
      ServerTimestamp: raknet.GetTimeMilliseconds(),
    }
    if err = this.SendPackage(toEncapsulate); err != nil {
      log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
      return
    }
  case mcpe.ID_MCPE_BATCH:
    this.handleMcpeBatch(pktData)
  case mcpe.ID_MCPE_LOGIN:
    lp := new(mcpe.MCPELogin)
    err := lp.Decode(bytes.NewReader(pktData))
    if err != nil {
      log.Printf("Error whilst handling message from %s", this.endpoint.String(), err)
      return
    }

    //this.AbandonWithReason("Hello! You're being disconnected because I didn't implement proxying!")
    log.Println("Log in successful, attempting a connection now...")
    this.loginPkt = lp
    this.Connect(&Server{
      Address: &net.UDPAddr{
        IP: net.IPv4(127, 0, 0, 1),
        Port: 19134,
      },
      Name: "test",
    })
  default:
    log.Printf("Unknown packet: %s", hex.EncodeToString(pktBytes))
  }
}

func (this *Session) handleConnected(pktBytes []byte) {
  // Perform serverbound message rewriting, if required.
  // TODO: Handle rewriting.

  // Forward the message on.
  this.serverConnection.packetQueue <- pktBytes
}
