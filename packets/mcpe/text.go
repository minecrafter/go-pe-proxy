package mcpe

import (
	"../raknet"
	"io"
)

type MCPEText struct {
	Type    MCPETextType
	Sender  string
	Message string
}

type MCPETextType byte

const (
	TEXT_TYPE_RAW MCPETextType = iota
	TEXT_TYPE_CHAT
	TEXT_TYPE_TRANSLATION
	TEXT_TYPE_POPUP
	TEXT_TYPE_TIP
)

func (*MCPEText) Id() byte {
	return ID_MCPE_TEXT
}

func (pkt *MCPEText) Decode(reader io.Reader) (err error) {
	tb, err := raknet.ReadByte(reader)
	if err != nil {
		return
	}

	rt := MCPETextType(tb)

	if rt == TEXT_TYPE_CHAT {
		s, e := raknet.ReadString(reader)
		pkt.Sender = s

		if err != nil {
			return e
		}
	}

	m, err := raknet.ReadString(reader)
	if err != nil {
		return
	}

	pkt.Type = rt
	pkt.Message = m
	return
}

func (pkt MCPEText) Encode(writer io.Writer) (err error) {
	err = raknet.WriteByte(writer, ID_MCPE_TEXT)
	if err != nil {
		return
	}
	err = raknet.WriteByte(writer, byte(pkt.Type))
	if err != nil {
		return
	}

	if pkt.Type == TEXT_TYPE_CHAT {
		err = raknet.WriteString(writer, pkt.Sender)
		if err != nil {
			return
		}
	}

	err = raknet.WriteString(writer, pkt.Message)
	if err != nil {
		return
	}
	return
}
