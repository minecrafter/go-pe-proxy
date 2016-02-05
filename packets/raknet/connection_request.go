package raknet

import (
  "io"
)

type RakNetConnectionRequest struct {
  GUID int64
  Timestamp int64
  Security byte
}

func (pkt RakNetConnectionRequest) Id() byte {
  return ID_CONNECTION_REQUEST
}

func (pkt *RakNetConnectionRequest) Decode(reader io.Reader) (err error) {
  guid, err := ReadInt64(reader)
  ts, err := ReadInt64(reader)
  secByte, err := ReadByte(reader)

  pkt.GUID = guid
  pkt.Timestamp = ts
  pkt.Security = secByte
  return
}

func (pkt RakNetConnectionRequest) Encode(writer io.Writer) (err error) {
  err = WriteByte(writer, ID_CONNECTION_REQUEST)
  if err != nil {
    return
  }
  err = WriteInt64(writer, pkt.GUID)
  if err != nil {
    return
  }
  err = WriteInt64(writer, pkt.Timestamp)
  if err != nil {
    return
  }
  err = WriteByte(writer, pkt.Security)
  if err != nil {
    return
  }
  return
}
