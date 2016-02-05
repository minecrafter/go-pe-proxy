package raknet

import (
	"../../util"
	"bytes"
	"errors"
	"io"
)

type RakNetDatagram struct {
	Type                   byte
	DatagramSequenceNumber int32
	Payload                []*EncapsulatedPacketPart
}

// Creates datagrams to send. Assumes that each packet can be sent within the MTU.
// Splits as needed.
func CreateDatagrams(rn *util.AtomicInteger, dsn *util.AtomicInteger, pkt EncodablePacket, mtu int16) (*[]RakNetDatagram, error) {
	var datagrams []RakNetDatagram
	maxDataSize := mtu - 60 // Allows for some overhead

	// Create our first datagram.
	currentDatagram := RakNetDatagram{
		DatagramSequenceNumber: dsn.IncrementAndGet(),
	}
	var currentPayload []*EncapsulatedPacketPart
	var curSz = 0

	// EncapsulatePacket will perform splitting if this packet is too large.
	encapsulated, err := EncapsulatePacket(rn, pkt, maxDataSize)
	if err != nil {
		return nil, err
	}

	t := ID_DATA_4
	if len(*encapsulated) > 1 {
		t = ID_DATA_C // Packet was split.
	}
	currentDatagram.Type = t

	// Try to encode the encapsulated packets.
	thisPktBuf := new(bytes.Buffer)
	for _, p := range *encapsulated {
		err = p.Encode(thisPktBuf)
		if err != nil {
			return nil, err
		}

		// Reject this datagram if it is obviously too big.
		if thisPktBuf.Len() > int(maxDataSize) {
			return nil, errors.New("Datagram too big")
		}

		if curSz+thisPktBuf.Len() > int(maxDataSize) {
			// This datagram is full. Create a new one.
			currentDatagram.Payload = currentPayload
			datagrams = append(datagrams, currentDatagram)
			curSz = 0

			currentPayload = []*EncapsulatedPacketPart{&p}
			currentDatagram = RakNetDatagram{
				Type: t,
				DatagramSequenceNumber: dsn.IncrementAndGet(),
			}
		} else {
			currentPayload = append(currentPayload, &p)
			curSz = curSz + thisPktBuf.Len()
		}

		thisPktBuf.Reset()
	}

	// Clean up the mess
	currentDatagram.Payload = currentPayload
	datagrams = append(datagrams, currentDatagram)

	return &datagrams, nil
}

func (pkt *RakNetDatagram) Decode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	t, err := r.ReadByte()
	if err != nil {
		return err
	}

	pkt.Type = t

	dsn, err := ReadInt24(r)
	if err != nil {
		return err
	}

	pkt.DatagramSequenceNumber = dsn

	var eps []*EncapsulatedPacketPart

	for {
		ep := new(EncapsulatedPacketPart)
		err := ep.Decode(r)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		eps = append(eps, ep)
	}

	pkt.Payload = eps
	return nil
}

func (pkt RakNetDatagram) Encode(writer io.Writer) error {
	err := WriteByte(writer, pkt.Type)
	if err != nil {
		return err
	}
	err = WriteInt24(writer, pkt.DatagramSequenceNumber)
	if err != nil {
		return err
	}

	for _, item := range pkt.Payload {
		err = item.Encode(writer)
		if err != nil {
			return err
		}
	}

	return nil
}
