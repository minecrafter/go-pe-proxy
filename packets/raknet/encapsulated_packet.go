package raknet

import (
  "bytes"
  "fmt"
  "errors"
  "math"
  "math/rand"
  "io"
  "../../util"
)

type EncapsulatedPacketPart struct {
  Reliability Reliability

  ReliabilityNumber int32
  OrderingIndex int32
  OrderingChannel byte

  PartCount int32
  PartId int16
  PartIndex int32

  Payload []byte
}

func sliceBytes(in []byte, size int) [][]byte {
  total := int(math.Ceil(float64(len(in)) / float64(size)))
	var sliced [][]byte
	for i := 0; i < total; i++ {
		s := i * size
		e := (i + 1) * size
		if e > len(in) {
			e = len(in)
		}
		sliced = append(sliced, in[s:e])
	}
	return sliced
}

func reliabilityToFlags(r Reliability, split bool) (byte) {
  flags := byte(r << 5)
  if split {
    flags = flags | 16
  }
  return flags
}

func decodeFlags(flags byte) (Reliability, bool, bool) {
  wasSplitNum := flags & 16
  r := flags >> 5
  if r >= 0 && r <= 7 {
    return Reliability(r), wasSplitNum > 0, true
  }
  return Unreliable, false, false
}

func EncapsulatePacket(rn *util.AtomicInteger, pkt EncodablePacket, maxDataSize int16) (*[]EncapsulatedPacketPart, error) {
  buf := new(bytes.Buffer)
  err := pkt.Encode(buf)
  if err != nil {
    return nil, err
  }

  var pkts []EncapsulatedPacketPart
  split := sliceBytes(buf.Bytes(), int(maxDataSize))
  partId := int16(0)
  if len(split) > 1 {
    partId = int16(rand.Int31n(8000000)) // that ought to do it
  }
  for i, slice := range split {
    p := EncapsulatedPacketPart{
      Reliability: Reliable,
      ReliabilityNumber: rn.IncrementAndGet(),
      PartCount: int32(len(split)),
      PartIndex: int32(i),
      PartId: partId,
      Payload: slice,
    }
    pkts = append(pkts, p)
  }
  return &pkts, nil
}

func (ep *EncapsulatedPacketPart) Decode(buf *bytes.Buffer) (error) {
  // TODO: This function needs cleanup
  f, err := buf.ReadByte()
  if err != nil {
    return err
  }
  r, wasSplit, ok1 := decodeFlags(f)

  if !ok1 {
    return errors.New("Flags are invalid.")
  }

  ln, err := ReadInt16(buf)
  if err != nil {
    return err
  }

  if r == Reliable || r == ReliableOrdered ||
    r == ReliableSequenced || r == ReliableWithAckReceipt ||
    r == ReliableOrderedWithAckReceipt {
    rn, err := ReadInt24(buf)
    if err != nil {
      return err
    }
    ep.ReliabilityNumber = rn
  }

  if r == UnreliableSequenced || r == ReliableOrdered || r == ReliableSequenced ||
    r == ReliableOrderedWithAckReceipt {
    oi, err := ReadInt24(buf)
    if err != nil {
      return err
    }
    oc, err := buf.ReadByte()
    if err != nil {
      return err
    }
    ep.OrderingIndex = oi
    ep.OrderingChannel = oc
  }

  if wasSplit {
    pc, err := ReadInt32(buf)
    if err != nil {
      return err
    }
    pi, err := ReadInt16(buf)
    if err != nil {
      return err
    }
    pix, err := ReadInt32(buf)
    if err != nil {
      return err
    }

    ep.PartCount = pc
    ep.PartId = pi
    ep.PartIndex = pix
  }

  ep.Reliability = r
  l := int(math.Ceil(float64(ln) / float64(8)))

  if l < 0 {
    // Prevent crashes
    return fmt.Errorf("Invalid packet length %d", l)
  }

  b := make([]byte, l)
  _, err = buf.Read(b)
  ep.Payload = b

  return err
}

func (ep EncapsulatedPacketPart) Encode(writer io.Writer) (err error) {
  flags := reliabilityToFlags(ep.Reliability, ep.PartCount > 1)

  // Write the packet
  err = WriteByte(writer, flags)
  err = WriteInt16(writer, int16(len(ep.Payload) * 8))
  if ep.Reliability == Reliable || ep.Reliability == ReliableOrdered ||
    ep.Reliability == ReliableSequenced || ep.Reliability == ReliableWithAckReceipt ||
    ep.Reliability == ReliableOrderedWithAckReceipt {
    err = WriteInt24(writer, ep.ReliabilityNumber)
  }

  if ep.Reliability == UnreliableSequenced || ep.Reliability == ReliableOrdered ||
    ep.Reliability == ReliableSequenced || ep.Reliability == ReliableOrderedWithAckReceipt {
    err = WriteInt24(writer, ep.OrderingIndex)
    err = WriteByte(writer, ep.OrderingChannel)
  }

  if ep.PartCount > 1 {
    err = WriteInt32(writer, ep.PartCount)
    err = WriteInt16(writer, ep.PartId)
    err = WriteInt32(writer, ep.PartIndex)
  }

  _, err = writer.Write(ep.Payload)
  return
}
