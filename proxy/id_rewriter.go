package proxy

import (
  "encoding/binary"
  "../packets/mcpe"
)

type EntityIdRewriter struct {
  clientId int64
  serverId int64
}

func NewEntityIdRewriter(initialId int64) EntityIdRewriter {
  return EntityIdRewriter{initialId, initialId}
}

func (rw *EntityIdRewriter) SetNewServerId(id int64) {
  rw.serverId = id
}

// Rewrites a slice of bytes for packets we don't want to process.
func (rw EntityIdRewriter) RewriteBytes(packet []byte) []byte {
  switch packet[0] {
  case mcpe.ID_MCPE_ANIMATE:
    i := binary.BigEndian.Uint64(packet[1:])
    if i == uint64(rw.serverId) {
      binary.BigEndian.PutUint64(packet[1:], uint64(rw.clientId))
    }
    return packet
  }
  return packet
}

// Rewrites entity IDs for packets we are processing.
func (rw EntityIdRewriter) RewritePacket(packet interface{}) interface{} {
  switch pkt := packet.(type) {
  case mcpe.MCPEStartGame:
    if pkt.EntityId == rw.serverId {
      pkt.EntityId = rw.clientId
    }
    return pkt
  default:
    return pkt
  }
}
