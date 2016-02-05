package mcpe

import (
	"../raknet"
	"io"
)

type MCPEStartGame struct {
	Seed      int32
	Dimension byte
	Generator int32
	Gamemode  int32
	EntityId  int64
	Spawn     BlockCoordinates
	Location  PlayerLocation
	Unknown   byte
}

func (*MCPEStartGame) Id() byte {
	return ID_MCPE_START_GAME
}

func (pkt *MCPEStartGame) Decode(reader io.Reader) (err error) {
	seed, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}
	dim, err := raknet.ReadByte(reader)
	if err != nil {
		return
	}
	gen, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}
	gm, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}
	eid, err := raknet.ReadInt64(reader)
	if err != nil {
		return
	}
	s, err := NewBlockCoordinates(reader)
	if err != nil {
		return
	}
	p, err := NewPlayerLocation(reader)
	if err != nil {
		return
	}
	u, err := raknet.ReadByte(reader)
	if err != nil {
		return
	}

	pkt.Seed = seed
	pkt.Dimension = dim
	pkt.Generator = gen
	pkt.Gamemode = gm
	pkt.EntityId = eid
	pkt.Spawn = s
	pkt.Location = p
	pkt.Unknown = u
	return
}

func (pkt MCPEStartGame) Encode(writer io.Writer) (err error) {
	err = raknet.WriteInt32(writer, pkt.Seed)
	if err != nil {
		return
	}
	err = raknet.WriteByte(writer, pkt.Dimension)
	if err != nil {
		return
	}
	err = raknet.WriteInt32(writer, pkt.Generator)
	if err != nil {
		return
	}
	err = raknet.WriteInt32(writer, pkt.Gamemode)
	if err != nil {
		return
	}
	err = raknet.WriteInt64(writer, pkt.EntityId)
	if err != nil {
		return
	}
	err = pkt.Spawn.Write(writer)
	if err != nil {
		return
	}
	err = pkt.Location.Write(writer)
	if err != nil {
		return
	}
	err = raknet.WriteByte(writer, pkt.Unknown)
	if err != nil {
		return
	}
	return
}
