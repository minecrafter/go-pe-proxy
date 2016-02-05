package proxy

import (
  "bytes"
  "encoding/hex"
  "fmt"
  "net"
  "log"
  "../packets/raknet"
)

type unknownSessionEntry struct {
  buf []byte
  endpoint *net.UDPAddr
}

type unknownSession struct {
  proxy *Proxy

  // INTERNAL: channel used to send packets for processing
  processQueue chan unknownSessionEntry
}

func NewUnknownSession(proxy *Proxy) *unknownSession {
  return &unknownSession{
    proxy: proxy,
    processQueue: make(chan unknownSessionEntry, 20000),
  }
}

func (this *unknownSession) Process() {
  for item := range this.processQueue {
    in := item.buf
    endpoint := item.endpoint
    pktData := in[1:]

    switch in[0] {
    case raknet.ID_UNCONNECTED_PING:
      pkt := new(raknet.RakNetUnconnectedPing)
      err := pkt.Decode(bytes.NewReader(pktData))
      if err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }

      log.Printf("Handling an unconnected ping packet from %s.", endpoint.String())
      name := fmt.Sprintf("MCPE;Test;38;0.13.0;%d;25000", this.proxy.Registry.Len())
      reply := raknet.NewRakNetUnconnectedPong(pkt.PingId, this.proxy.guid, name)
      if err = raknet.WriteUDP(this.proxy.conn, endpoint, reply); err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }
      break
    case raknet.ID_OPEN_CONNECTION_REQUEST_1:
      pkt := new(raknet.RakNetOpenConnectionRequest1)
      err := pkt.Decode(bytes.NewReader(pktData))
      if err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }

      // Go figure. The packet's full size is (almost) the exact MTU.
      realMtu := len(item.buf) + 32

      log.Printf("Handling the first stage request packet from %s", endpoint.String())
      reply := raknet.NewRakNetOpenConnectionReply1(this.proxy.guid, 0, int16(realMtu))
      if err = raknet.WriteUDP(this.proxy.conn, endpoint, reply); err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }
      break
    case raknet.ID_OPEN_CONNECTION_REQUEST_2:
      pkt := new(raknet.RakNetOpenConnectionRequest2)
      err := pkt.Decode(bytes.NewReader(pktData))
      if err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }

      // Create the connection for this client.
      connection := NewSession(this.proxy, pkt.MTU, endpoint)
      if ok := this.proxy.Registry.Register(connection); !ok {
        continue
      }

      // Initialize the client:
      go connection.handleSession()

      log.Printf("Created a connection for %s.", endpoint.String())

      // Send response. Welcome to the club!
      reply := raknet.NewRakNetOpenConnectionReply2(this.proxy.guid, *endpoint, pkt.MTU)
      if err = raknet.WriteUDP(this.proxy.conn, endpoint, reply); err != nil {
        log.Printf("Error whilst handling message from %s", endpoint.String(), err)
        continue
      }
      break
    default:
      log.Printf("Unknown packet from %s. Packet ID: %d, Hex dump: %s", endpoint.String(), in[0], hex.EncodeToString(pktData))
      break
    }

    continue
  }
}
