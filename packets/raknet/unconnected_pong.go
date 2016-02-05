package raknet

import (
	"io"
)

type RakNetUnconnectedPong struct {
	PingId   int64
	ServerId int64
	Name     string
}

func (pkt RakNetUnconnectedPong) Id() byte {
	return ID_UNCONNECTED_PONG
}

func NewRakNetUnconnectedPong(pingId int64, serverId int64, name string) RakNetUnconnectedPong {
	return RakNetUnconnectedPong{
		PingId:   pingId,
		ServerId: serverId,
		Name:     name,
	}
}

func (pkt RakNetUnconnectedPong) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_UNCONNECTED_PONG)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.PingId)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.ServerId)
	if err != nil {
		return
	}
	_, err = writer.Write(MAGIC)
	if err != nil {
		return
	}
	err = WriteString(writer, pkt.Name)
	return
}
