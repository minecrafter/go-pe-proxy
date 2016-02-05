package mcpe

import (
	"../raknet"
	"errors"
	"github.com/pborman/uuid"
	"io"
)

// It has two actions, which have very different structure.
type PlayerListAction byte

const (
	PlayerListAdd    PlayerListAction = 0
	PlayerListRemove PlayerListAction = 1
)

type MCPEPlayerListPlayer struct {
	UUID     uuid.UUID
	EntityId int64
	Username string
	Skin     Skin
}

type MCPEPlayerList struct {
	Action  PlayerListAction
	Players []MCPEPlayerListPlayer
}

func (*MCPEPlayerList) Id() byte {
	return ID_MCPE_PLAYER_LIST
}

func (pkt MCPEPlayerList) Encode(writer io.Writer) (err error) {
	err = raknet.WriteByte(writer, ID_MCPE_PLAYER_LIST)
	if err != nil {
		return
	}
	err = raknet.WriteByte(writer, byte(pkt.Action))
	if err != nil {
		return
	}
	err = raknet.WriteInt32(writer, int32(len(pkt.Players)))
	if err != nil {
		return
	}

	switch pkt.Action {
	case PlayerListAdd:
		for _, item := range pkt.Players {
			err = WriteUUID(writer, item.UUID)
			if err != nil {
				return err
			}
			err = raknet.WriteInt64(writer, item.EntityId)
			if err != nil {
				return err
			}
			err = raknet.WriteString(writer, item.Username)
			if err != nil {
				return err
			}
			err = item.Skin.Encode(writer)
			if err != nil {
				return err
			}
		}
	case PlayerListRemove:
		for _, item := range pkt.Players {
			err = WriteUUID(writer, item.UUID)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Invalid action")
	}

	return
}

func (pkt *MCPEPlayerList) Decode(reader io.Reader) (err error) {
	a, err := raknet.ReadByte(reader)
	if err != nil {
		return
	}

	ln, err := raknet.ReadInt32(reader)
	if err != nil {
		return
	}

	switch PlayerListAction(a) {
	case PlayerListAdd:
		for i := 0; i < int(ln); i++ {
			uuid, err := ReadUUID(reader)
			if err != nil {
				return err
			}
			eid, err := raknet.ReadInt64(reader)
			if err != nil {
				return err
			}
			n, err := raknet.ReadString(reader)
			if err != nil {
				return err
			}
			s, err := NewSkin(reader)
			if err != nil {
				return err
			}

			pkt.Players = append(pkt.Players, MCPEPlayerListPlayer{
				UUID:     uuid,
				EntityId: eid,
				Username: n,
				Skin:     *s,
			})
		}
	case PlayerListRemove:
		for i := 0; i < int(ln); i++ {
			uuid, err := ReadUUID(reader)
			if err != nil {
				return err
			}

			pkt.Players = append(pkt.Players, MCPEPlayerListPlayer{
				UUID: uuid,
			})
		}
	default:
		return errors.New("Invalid action")
	}

	return
}
