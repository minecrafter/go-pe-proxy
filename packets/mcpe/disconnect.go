package mcpe

import (
	"../raknet"
	"io"
)

type MCPEDisconnect struct {
	Message string
}

func (*MCPEDisconnect) Id() byte {
	return ID_MCPE_DISCONNECT
}

func (pkt *MCPEDisconnect) Decode(reader io.Reader) (err error) {
	msg, err := raknet.ReadString(reader)
	if err != nil {
		return
	}
	pkt.Message = msg
	return
}

func (pkt MCPEDisconnect) Encode(writer io.Writer) (err error) {
	err = raknet.WriteByte(writer, ID_MCPE_DISCONNECT)
	if err != nil {
		return
	}
	err = raknet.WriteString(writer, pkt.Message)
	return
}
