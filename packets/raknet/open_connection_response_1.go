package raknet

import (
  "bytes"
  "errors"
  "io"
)

type RakNetOpenConnectionReply1 struct {
  GUID int64
  Secure byte
  MTU int16
}

func NewRakNetOpenConnectionReply1(guid int64, secure byte, mtu int16) RakNetOpenConnectionReply1 {
  return RakNetOpenConnectionReply1{
    GUID: guid,
    Secure: secure,
    MTU: mtu,
  }
}

func (pkt RakNetOpenConnectionReply1) Id() byte {
  return ID_OPEN_CONNECTION_REPLY_1
}

func (pkt *RakNetOpenConnectionReply1) Decode(reader io.Reader) (err error) {
  data := make([]byte, 16)
  _, err = reader.Read(data)

  if !bytes.Equal(data, MAGIC) {
    return errors.New("Offline magic not valid.")
  }

  guid, err := ReadInt64(reader)
  if err != nil {
    return
  }
  secure, err := ReadByte(reader)
  if err != nil {
    return
  }
  mtu, err := ReadInt16(reader)
  if err != nil {
    return
  }

  pkt.GUID = guid
  pkt.Secure = secure
  pkt.MTU = mtu

  return
}

func (pkt RakNetOpenConnectionReply1) Encode(writer io.Writer) (err error) {
  err = WriteByte(writer, ID_OPEN_CONNECTION_REPLY_1)
  if err != nil {
    return
  }
  _, err = writer.Write(MAGIC)
  if err != nil {
    return
  }
  err = WriteInt64(writer, pkt.GUID)
  if err != nil {
    return
  }
  err = WriteByte(writer, pkt.Secure)
  if err != nil {
    return
  }
  err = WriteInt16(writer, pkt.MTU)
  if err != nil {
    return
  }
  return
}
