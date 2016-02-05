package mcpe

import (
  "github.com/pborman/uuid"
  "io"
  "../raknet"
)

type MCPELogin struct {
  Username string
  Protocol1 int32
  Protocol2 int32
  ClientGuid int64
  ClientUuid uuid.UUID
  ServerAddress string
  ClientSecret string
  Skin *Skin
}

func (pkt *MCPELogin) Id() byte {
  return ID_MCPE_LOGIN
}

func (pkt *MCPELogin) Decode(reader io.Reader) (err error) {
  username, err := raknet.ReadString(reader)
  if err != nil {
    return err
  }

  p1, err := raknet.ReadInt32(reader)
  if err != nil {
    return err
  }
  p2, err := raknet.ReadInt32(reader)
  if err != nil {
    return err
  }

  guid, err := raknet.ReadInt64(reader)
  if err != nil {
    return err
  }
  uuid, err := ReadUUID(reader)
  if err != nil {
    return err
  }

  serverAddress, err := raknet.ReadString(reader)
  if err != nil {
    return err
  }

  secret, err := raknet.ReadString(reader)
  if err != nil {
    return err
  }

  skin, err := NewSkin(reader)
  if err != nil {
    return err
  }

  pkt.Username = username
  pkt.Protocol1 = p1
  pkt.Protocol2 = p2
  pkt.ClientGuid = guid
  pkt.ClientUuid = uuid
  pkt.ServerAddress = serverAddress
  pkt.ClientSecret = secret
  pkt.Skin = skin
  return nil
}

func (pkt *MCPELogin) Encode(writer io.Writer) (err error) {
  err = raknet.WriteByte(writer, ID_MCPE_LOGIN)
  if err != nil {
    return
  }
  err = raknet.WriteString(writer, pkt.Username)
  if err != nil {
    return err
  }
  err = raknet.WriteInt32(writer, pkt.Protocol1)
  if err != nil {
    return err
  }
  err = raknet.WriteInt32(writer, pkt.Protocol2)
  if err != nil {
    return err
  }
  err = raknet.WriteInt64(writer, pkt.ClientGuid)
  if err != nil {
    return err
  }
  err = WriteUUID(writer, pkt.ClientUuid)
  if err != nil {
    return err
  }
  err = raknet.WriteString(writer, pkt.ServerAddress)
  if err != nil {
    return err
  }
  err = raknet.WriteString(writer, pkt.ClientSecret)
  if err != nil {
    return err
  }
  err = pkt.Skin.Encode(writer)
  return
}
