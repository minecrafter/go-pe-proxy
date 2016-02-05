package mcpe

import (
  "io"
  "../raknet"
)

type Skin struct {
  Slim bool
  Alpha byte
  Data []byte
}

func (skin Skin) Encode(writer io.Writer) (error) {
  err := raknet.WriteByte(writer, skin.Alpha)
  if err != nil {
    return err
  }

  err = raknet.WriteBoolean(writer, skin.Slim)
  if err != nil {
    return err
  }

  err = raknet.WriteInt16(writer, int16(len(skin.Data)))
  if err != nil {
    return err
  }

  _, err = writer.Write(skin.Data)
  if err != nil {
    return err
  }

  return nil
}

func NewSkin(reader io.Reader) (*Skin, error) {
  alpha, err := raknet.ReadByte(reader)
  if err != nil {
    return nil, err
  }

  slim, err := raknet.ReadBoolean(reader)
  if err != nil {
    return nil, err
  }

  dataLn, err := raknet.ReadInt16(reader)
  if err != nil {
    return nil, err
  }

  data := make([]byte, dataLn)
  _, err = reader.Read(data)
  if err != nil {
    return nil, err
  }

  return &Skin{slim, alpha, data}, nil
}
