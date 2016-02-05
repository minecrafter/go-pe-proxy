package mcpe

import (
  "../raknet"
  "io"
)

type MCPEPlayerStatus struct {
  Status int32
}

func (*MCPEPlayerStatus) Id() byte {
  return ID_MCPE_PLAYER_STATUS
}

func (pkt *MCPEPlayerStatus) Decode(reader io.Reader) (err error) {
  status, err := raknet.ReadInt32(reader)
  if err != nil {
    return
  }
  pkt.Status = status
  return
}

func (pkt MCPEPlayerStatus) Encode(writer io.Writer) (err error) {
  err = raknet.WriteByte(writer, ID_MCPE_PLAYER_STATUS)
  if err != nil {
    return
  }
  err = raknet.WriteInt32(writer, pkt.Status)
  return
}
