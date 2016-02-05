package raknet

import (
	"io"
	"sort"
)

type RakNetAck struct {
	Acknowledged []Range
}

type Range struct {
	Min int
	Max int
}

func (RakNetAck) Id() byte {
	return ID_ACK
}

func SliceAck(acks []int) []Range {
	sort.Ints(acks)

	sliced := make([]Range, 0)

	start := acks[0]
	currSeq := start

	for _, item := range acks[1:] {
		diff := item - currSeq
		if diff == 1 {
			// Number is sequential, update currSeq and continue
			currSeq = item
			continue
		}

		// It's not sequential, so we have a range.
		r := Range{start, currSeq}
		sliced = append(sliced, r)

		// Reset start and currSeq
		start = item
		currSeq = item
	}

	// Anything left over needs to be cleaned up.
	sliced = append(sliced, Range{start, currSeq})

	return sliced
}

func (pkt RakNetAck) Encode(writer io.Writer) (err error) {
	err = WriteByte(writer, ID_ACK)

	// Encode them
	err = WriteInt16(writer, int16(len(pkt.Acknowledged)))

	for _, item := range pkt.Acknowledged {
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

func (pkt RakNetAck) Decode(reader io.Reader) (err error) {
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

	pkt.Acknowledged = sliced
	return err
}
