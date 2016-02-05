package raknet

import (
	"io"
	"net"
)

type RakNetConnectionRequestAccepted struct {
	SystemAddress     net.UDPAddr
	IncomingTimestamp int64
	ServerTimestamp   int64
}

func (pkt RakNetConnectionRequestAccepted) Id() byte {
	return ID_CONNECTION_REQUEST_ACCEPTED
}

func (pkt *RakNetConnectionRequestAccepted) Decode(reader io.Reader) (err error) {
	address, err := ReadUDPAddr(reader)
	if err != nil {
		return
	}

	// Read in 10 addresses, and ignore them. What. The. Fuck.
	for i := 0; i < 10; i++ {
		ReadUDPAddr(reader)
	}

	ts1, err := ReadInt64(reader)
	if err != nil {
		return
	}
	ts2, err := ReadInt64(reader)
	if err != nil {
		return
	}

	pkt.SystemAddress = address
	pkt.IncomingTimestamp = ts1
	pkt.ServerTimestamp = ts2
	return
}

func (pkt RakNetConnectionRequestAccepted) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_CONNECTION_REQUEST_ACCEPTED)
	if err != nil {
		return
	}
	err = WriteUDPAddr(writer, pkt.SystemAddress)
	if err != nil {
		return
	}
	// 10 IPs. Yes, I did say that.
	dummyIp1 := net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 19132,
	}
	dummyIp2 := net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 19132,
	}
	err = WriteUDPAddr(writer, dummyIp1)
	if err != nil {
		return
	}
	for i := 0; i < 9; i++ {
		err = WriteUDPAddr(writer, dummyIp2)
		if err != nil {
			return
		}
	}
	err = WriteInt64(writer, pkt.IncomingTimestamp)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.ServerTimestamp)
	if err != nil {
		return
	}
	return
}
