package raknet

import (
  "bytes"
  "io"
)

// Generic class used when we don't know the structure of the package.
type GenericRakNetPackage struct {
  PacketId byte
  Payload []byte
}

func (pkt GenericRakNetPackage) Id() byte {
  return pkt.PacketId
}

func (pkt *GenericRakNetPackage) Decode(reader io.Reader) (err error) {
  // Drain the reader
  b := new(bytes.Buffer)
  _, err = io.Copy(b, reader)

  pkt.Payload = b.Bytes()
  return
}

func (pkt GenericRakNetPackage) Encode(writer io.Writer) (err error) {
  err = WriteByte(writer, pkt.PacketId)
  _, err = writer.Write(pkt.Payload)
  return
}
