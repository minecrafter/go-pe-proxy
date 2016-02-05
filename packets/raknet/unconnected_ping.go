package raknet

import (
	"io"
)

type RakNetUnconnectedPing struct {
	PingId int64
}

func (pkt RakNetUnconnectedPing) Id() byte {
	return ID_UNCONNECTED_PING
}

func (pkt *RakNetUnconnectedPing) Decode(reader io.Reader) (err error) {
	id, err := ReadInt64(reader)
	pkt.PingId = id
	return
}

func (pkt RakNetUnconnectedPing) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_UNCONNECTED_PING)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.PingId)
	if err != nil {
		return
	}
	_, err = writer.Write(MAGIC)
	return
}
