package client

import (
  "bytes"
  "encoding/hex"
  "math/rand"
  "net"
  "log"
  "../player"
  "../packets/raknet"
  "../util"
  "time"
)

type BackendConnection struct {
  Server Server
  Player player.Connection
  Connection *net.UDPConn

  guid int64
  mtu int16
  pingTicker *time.Ticker
  ReliabilityNumber *util.AtomicInteger

  lastTouched time.Time
}

type BackendConnectionHandler interface {
  OnMessage(buf []byte)
  OnClose()
}

func CreateBackendConnection(player player.Connection, server Server) BackendConnection {
  return BackendConnection{
    player: player,
    server: server,
    guid: rand.Int63(),
    reliabilityNumber: new(util.AtomicInteger),
  }
}

func (bc *BackendConnection) Close() {
  if bc.connection != nil {
    // TODO: Check state and send a remote message.
    bc.connection.Close()
  }
}

func (bc *BackendConnection) InitiateConnection() {
  // Dial the server address.
  conn, err := net.DialUDP("udp4", nil, &bc.server.Address)
  if err != nil {
    bc.player.connectionFailure(bc.server, err)
    return
  }

  // Begin RakNet handshake process.
  bc.connection = conn
  go bc.handleConnection()

  // Send our first packet.
  pkt := raknet.RakNetOpenConnectionRequest1{
    ProtocolVersion: 7,
  }
  if err = bc.SendPacket(pkt); err != nil {
    bc.player.connectionFailure(bc.server, err)
    return
  }
}

func (bc *BackendConnection) SendPacket(pkt raknet.EncodablePacket) (error) {
  return raknet.WriteUDP(bc.connection, &bc.server.Address, pkt)
}

func (bc *BackendConnection) sendPingPacket() {
  pkt := raknet.NewConnectedPingWithCurrentTime()
  if err := bc.SendPacket(pkt); err != nil {
    log.Printf("Error whilst writing ping to %s", (&bc.server.Address).String(), err)
  }
}

func (bc *BackendConnection) TouchUpdated() {
  bc.lastTouched = time.Now()
}

func (bc *BackendConnection) connectionTask() {
  defer bc.connection.Close()

  for {
    buf := make([]byte, 1447) // TODO: Is this ok?
    _, endpoint, err := bc.connection.ReadFromUDP(buf)

    if err != nil {
      log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
      return
    }

    bc.handleIncoming(buf, endpoint)
  }
}
