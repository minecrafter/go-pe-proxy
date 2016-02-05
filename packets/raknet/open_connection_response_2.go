package raknet

import (
  "bytes"
  "errors"
  "net"
  "io"
)

type RakNetOpenConnectionReply2 struct {
  GUID int64
  ClientEndpoint net.UDPAddr
  MTU int16
}

func NewRakNetOpenConnectionReply2(guid int64, endpoint net.UDPAddr, mtu int16) RakNetOpenConnectionReply2 {
  return RakNetOpenConnectionReply2{
    GUID: guid,
    ClientEndpoint: endpoint,
    MTU: mtu,
  }
}

func (pkt RakNetOpenConnectionReply2) Id() byte {
  return ID_OPEN_CONNECTION_REPLY_2
}

func (pkt *RakNetOpenConnectionReply2) Decode(reader io.Reader) (err error) {
  data := make([]byte, 16)
  _, err = reader.Read(data)

  if !bytes.Equal(data, MAGIC) {
    return errors.New("Offline magic not valid.")
  }

  guid, err := ReadInt64(reader)
  addr, err := ReadUDPAddr(reader)
  mtu, err := ReadInt16(reader)

  pkt.GUID = guid
  pkt.ClientEndpoint = addr
  pkt.MTU = mtu

  return
}

func (pkt RakNetOpenConnectionReply2) Encode(writer io.Writer) (err error) {
  err = WriteByte(writer, ID_OPEN_CONNECTION_REPLY_2)
  _, err = writer.Write(MAGIC)
  err = WriteInt64(writer, pkt.GUID)
  err = WriteUDPAddr(writer, pkt.ClientEndpoint)
  err = WriteInt16(writer, pkt.MTU)
  err = WriteByte(writer, 0) // one-byte empty
  return
}
