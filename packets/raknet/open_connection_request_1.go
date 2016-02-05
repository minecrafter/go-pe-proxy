package raknet

import (
	"bytes"
	"errors"
	"io"
)

type RakNetOpenConnectionRequest1 struct {
	ProtocolVersion byte
	MTUFill         int16
}

func (pkt RakNetOpenConnectionRequest1) Id() byte {
	return ID_OPEN_CONNECTION_REQUEST_1
}

func (pkt *RakNetOpenConnectionRequest1) Decode(reader io.Reader) (err error) {
	data := make([]byte, 16)
	_, err = reader.Read(data)

	if !bytes.Equal(data, MAGIC) {
		return errors.New("Offline magic not valid.")
	}

	v, err := ReadByte(reader)
	pkt.ProtocolVersion = v

	return
}

func (pkt RakNetOpenConnectionRequest1) Encode(writer io.Writer) (err error) {
	// Ideally, all Encode functions should look like this.
	data := make([]byte, pkt.MTUFill)
	data[0] = ID_OPEN_CONNECTION_REQUEST_1
	copy(data[1:16], MAGIC)
	data[17] = pkt.ProtocolVersion
	_, err = writer.Write(data)
	return
}
