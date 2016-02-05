package raknet

import (
	"bytes"
	"errors"
	"io"
	"net"
)

type RakNetOpenConnectionRequest2 struct {
	ClientEndpoint *net.UDPAddr
	MTU            int16
	GUID           int64
}

func (pkt RakNetOpenConnectionRequest2) Id() byte {
	return ID_OPEN_CONNECTION_REQUEST_2
}

func (pkt *RakNetOpenConnectionRequest2) Decode(reader io.Reader) (err error) {
	magic := make([]byte, 16)
	_, err = reader.Read(magic)

	if !bytes.Equal(magic, MAGIC) {
		return errors.New("Offline magic not valid.")
	}

	addr, err := ReadUDPAddr(reader)
	if err != nil {
		return
	}
	mtu, err := ReadInt16(reader)
	if err != nil {
		return
	}
	guid, err := ReadInt64(reader)
	if err != nil {
		return
	}

	pkt.ClientEndpoint = &addr
	pkt.MTU = mtu
	pkt.GUID = guid

	return
}

func (pkt RakNetOpenConnectionRequest2) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_OPEN_CONNECTION_REQUEST_2)
	if err != nil {
		return
	}
	_, err = writer.Write(MAGIC)
	if err != nil {
		return
	}
	err = WriteUDPAddr(writer, *pkt.ClientEndpoint)
	if err != nil {
		return
	}
	err = WriteInt16(writer, pkt.MTU)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.GUID)
	return
}
