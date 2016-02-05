package raknet

import (
  "io"
)

type RakNetConnectedPong struct {
  Timestamp1 int64
  Timestamp2 int64
}

func (pkt RakNetConnectedPong) Id() byte {
  return ID_CONNECTED_PONG
}

func (pkt RakNetConnectedPong) Decode(reader io.Reader) (err error) {
  ts, err := ReadInt64(reader)
  pkt.Timestamp1 = ts
  ts2, err := ReadInt64(reader)
  pkt.Timestamp2 = ts2

  return
}

func (pkt RakNetConnectedPong) Encode(writer io.Writer) (err error) {
  err = WriteByte(writer, ID_CONNECTED_PONG)
  err = WriteInt64(writer, pkt.Timestamp1)
  err = WriteInt64(writer, pkt.Timestamp2)
  return
}
