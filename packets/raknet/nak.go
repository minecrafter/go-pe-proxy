package raknet

import (
	"io"
)

type RakNetNak struct {
	NotAcknowledged []Range
}

func (RakNetNak) Id() byte {
	return ID_NAK
}

func (pkt RakNetNak) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_ACK)

	err = WriteInt16(writer, int16(len(pkt.NotAcknowledged)))

	// Encode them
	for _, item := range pkt.NotAcknowledged {
		if item.Min != item.Max {
			err = WriteByte(writer, 0)
			err = WriteInt24(writer, int32(item.Min))
			err = WriteInt24(writer, int32(item.Max))
		} else {
			err = WriteByte(writer, 1)
			err = WriteInt24(writer, int32(item.Min))
		}
	}

	return
}

func (pkt RakNetNak) Decode(reader io.Reader) (err error) {
	sliced := make([]Range, 0)
	count, err := ReadInt16(reader)
	if err != nil {
		return
	}

	for i := 0; i < int(count); i++ {
		s, err := ReadByte(reader)
		if err != nil {
			return err
		}

		// Otherwise, decode a range
		if s == 1 {
			n, err := ReadInt24(reader)
			if err != nil {
				return err
			}
			sliced = append(sliced, Range{int(n), int(n)})
		} else {
			n1, err := ReadInt24(reader)
			if err != nil {
				return err
			}
			n2, err := ReadInt24(reader)
			if err != nil {
				return err
			}
			sliced = append(sliced, Range{int(n1), int(n2)})
		}
	}

	pkt.NotAcknowledged = sliced
	return
}
