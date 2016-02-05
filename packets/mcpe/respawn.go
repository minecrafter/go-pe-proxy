package mcpe

import (
	"../raknet"
	"io"
)

type MCPERespawn struct {
	Location PlayerLocation
}

func (*MCPERespawn) Id() byte {
	return ID_MCPE_RESPAWN
}

func (pkt *MCPERespawn) Decode(reader io.Reader) (err error) {
	l, err := NewPlayerLocation(reader)
	if err != nil {
		return
	}
	pkt.Location = l
	return
}

func (pkt MCPERespawn) Encode(writer io.Writer) (err error) {
	err = raknet.WriteByte(writer, ID_MCPE_RESPAWN)
	if err != nil {
		return
	}
	err = pkt.Location.Write(writer)
	return
}
