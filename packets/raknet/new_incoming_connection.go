package raknet

import (
	"io"
)

type RakNetNewIncomingConnection struct {
	Cookie   int32
	Secure   byte
	Port     int16
	Session1 int64
	Session2 int64
}

func (pkt RakNetNewIncomingConnection) Id() byte {
	return ID_NEW_INCOMING_CONNECTION
}

func (pkt RakNetNewIncomingConnection) Decode(reader io.Reader) (err error) {
	cookie, err := ReadInt32(reader)
	secure, err := ReadByte(reader)
	port, err := ReadInt16(reader)
	s1, err := ReadInt64(reader)
	s2, err := ReadInt64(reader)

	pkt.Cookie = cookie
	pkt.Secure = secure
	pkt.Port = port
	pkt.Session1 = s1
	pkt.Session2 = s2
	return
}

func (pkt RakNetNewIncomingConnection) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_NEW_INCOMING_CONNECTION)
	if err != nil {
		return
	}
	err = WriteInt32(writer, pkt.Cookie)
	if err != nil {
		return
	}
	err = WriteByte(writer, pkt.Secure)
	if err != nil {
		return
	}
	err = WriteInt16(writer, pkt.Port)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.Session1)
	if err != nil {
		return
	}
	err = WriteInt64(writer, pkt.Session2)
	return
}
