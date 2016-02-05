package mcpe

import (
  "encoding/binary"
  "io"
  "github.com/pborman/uuid"
)

func ReadFloat32(reader io.Reader) (val float32, err error) {
  err = binary.Read(reader, binary.BigEndian, &val)
  return
}

func WriteFloat32(writer io.Writer, val float32) (err error) {
  err = binary.Write(writer, binary.BigEndian, val)
  return
}

func ReadUUID(reader io.Reader) (val uuid.UUID, err error) {
  uuidBuf := make([]byte, 16)
  _, err = reader.Read(uuidBuf)
  if err != nil {
    return uuid.UUID(uuidBuf), err
  }
  val = uuid.UUID(uuidBuf)
  return
}

func WriteUUID(writer io.Writer, val uuid.UUID) (err error) {
  uuidBuf := []byte(val)
  _, err = writer.Write(uuidBuf)
  return
}
