package raknet

import (
	"io"
)

type RakNetConnectedPing struct {
	Timestamp int64
}

func (pkt RakNetConnectedPing) Id() byte {
	return ID_CONNECTED_PING
}

func NewConnectedPingWithCurrentTime() RakNetConnectedPing {
	return RakNetConnectedPing{GetTimeMilliseconds()}
}

func (pkt RakNetConnectedPing) Decode(reader io.Reader) (err error) {
	ts, err := ReadInt64(reader)
	pkt.Timestamp = ts
	return err
}

func (pkt RakNetConnectedPing) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_CONNECTED_PING)
	err = WriteInt64(writer, pkt.Timestamp)
	return
}
