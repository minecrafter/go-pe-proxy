package raknet

import (
	"io"
)

type Packet interface {
	Id() byte
}

type EncodablePacket interface {
	Encode(writer io.Writer) (err error)
}

type DecodablePacket interface {
	Decode(reader io.Reader) (err error)
}

type FullPacket interface {
	Packet
	EncodablePacket
	DecodablePacket
}
