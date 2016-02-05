package client

import (
  "encoding/hex"
  "../packets/mcpe"
  "../player"
  "../packets/raknet"
)

type BackendConnectedHandler struct {
  backendConnection proxy.BackendConnection
  player *proxy.Connection

  firstServer bool

  splitPackets raknet.SplitPacketHandler
}

func (h *BackendConnectedHandler) OnMessage(buf []byte) {
  switch buf[0] {
  case raknet.ID_CONNECTED_PONG:
    // Pong packet. Don't want to forward this to the remote client.
    pkt := new(raknet.RakNetConnectedPong)
    err := pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      log.Printf("Error whilst handling message from %s", endpoint.String(), err)
      return
    }

    log.Printf("Handling a connected pong packet from %s.", endpoint.String())
    bc.TouchUpdated()
    break
  case raknet.ID_DATA_4:
    // Datagram with encapsulated packet(s), not split.
    pkt := new(raknet.RakNetDatagram)
    err := pkt.Decode(in)
    if err != nil {
      log.Printf("Error whilst handling message from %s", endpoint.String(), err)
      return
    }

    for _, item := range pkt.Payload {
      log.Printf("Handling extracted packet: %s ...", hex.EncodeToString(item.Payload[0:15]))
      h.OnMessage(item.Payload)
    }
    break
  case mcpe.ID_MCPE_BATCH:
    // Batch packet. We have to handle this for certain packets.
    pkt := new(mcpe.MCPEBatch)
    err := pkt.Decode(bytes.NewReader(pktData))
    if err != nil {
      log.Printf("Error whilst handling message from %s", endpoint.String(), err)
      return
    }

    // Handle MCPEBatch packets in a separate function:
    h.OnBatchPacket(pkt)
    break
  default:
    // We don't know what this packet is, and we don't want to know. Forward it to
    // the client.
    log.Printf("Forwarding unknown packet %s to client %s.",
      hex.EncodeToString(item.Payload[0:15]), player.Endpoint.String())
    if _, err := h.player.Proxy.Listener.WriteToUDP(buf, h.player.Endpoint); err != nil {
      log.Printf("Unable to forward packet to %s: %s", h.player.Endpoint.String(), err.Error())
    }
  }
}

func (h *BackendConnectedHandler) OnBatchPacket(batchPacket *mcpe.MCPEBatch) {
  var forward [][]byte
  for _, payload := range batchPacket.Payload {
    log.Printf("Handling decompressed batch packet: %s ...", hex.EncodeToString(payload[0:15]))
    switch payload[0] {
    case ID_MCPE_PLAYER_STATUS:
      // Forward this packet iff this is the player's first server.
      if !h.firstServer {
        break
      }
      fallthrough
    case ID_MCPE_START_GAME:
      pkt := new(mcpe.MCPEStartGame)
      err := pkt.Decode(bytes.NewReader(payload))
      if err != nil {
        log.Printf("Error whilst handling message from %s", h.player.Endpoint.String(), err)
        return
      }

      // The first packet on which we act on in a major way.
      if h.firstServer {
        // This is our first server. Initialize the entity ID rewriter.
        // We don't need to rewrite this packet.
        rewriter := player.NewEntityIdRewriter(pkt.EntityId)
        h.player.EntityIdRewriter = rewriter
        // Forward this packet on to the client.
        forward = append(forward, payload)
      } else {
        // We don't want to write this packet to the client, but we want some
        // information from it.
        h.player.EntityIdRewriter.SetNewServerId(pkt.EntityId)

        // Send a respawn packet instead.
        spawnPacket := MCPERespawn{pkt.Location}
        var spBytes bytes.Buffer
        err = spawnPacket.Encode(spBytes)
        if err != nil {
          log.Printf("Error whilst handling message from %s", h.player.Endpoint.String(), err)
          return
        }
        forward = append(forward, spBytes.Bytes())
      }
    default:
      // Don't know what to do with this packet. Let's pack it up.
      forward = append(forward, payload)
    }
  }

  if len(forward > 0) {
    // Compress packets and forward them
    pkt := new(mcpe.MCPEBatch)
    pkt.Payload = forward
    if err := h.Connection.SendPackage(pkt); err != nil {
      log.Printf("Unable to forward packet to %s: %s", h.player.Endpoint.String(), err.Error())
    }
  }
}

func (h *BackendConnectedHandler) OnClose() {

}
