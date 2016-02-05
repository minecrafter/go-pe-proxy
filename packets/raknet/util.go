package raknet

import (
	"bytes"
	"encoding/binary"
	"github.com/kevinjos/openbci-golang-server/int24"
	"io"
	"net"
	"time"
)

// Useful time function.
func GetTimeMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// This function is needed since UDP is stateless.
func WriteUDP(conn *net.UDPConn, endpoint *net.UDPAddr, pkt EncodablePacket) (err error) {
	out := new(bytes.Buffer)
	err = pkt.Encode(out)
	if err != nil {
		return
	}

	_, err = conn.WriteToUDP(out.Bytes(), endpoint)
	return
}

// This function is needed since UDP is stateless.
func WriteUDPPreConnected(conn *net.UDPConn, pkt EncodablePacket) (err error) {
	out := new(bytes.Buffer)
	err = pkt.Encode(out)
	if err != nil {
		return
	}

	_, err = conn.Write(out.Bytes())
	return
}

func ReadString(reader io.Reader) (str string, err error) {
	ln, err := ReadUint16(reader)
	if err != nil {
		return "", err
	}
	asBytes := make([]byte, ln)
	_, err = reader.Read(asBytes)
	str = string(asBytes)
	return
}

func WriteString(writer io.Writer, str string) (err error) {
	err = WriteUint16(writer, uint16(len(str)))
	_, err = io.WriteString(writer, str)
	return
}

func ReadInt16(reader io.Reader) (val int16, err error) {
	t, err := ReadUint16(reader)
	return int16(t), err
}

func WriteInt16(writer io.Writer, val int16) (err error) {
	return WriteUint16(writer, uint16(val))
}

func ReadUint16(reader io.Reader) (val uint16, err error) {
	res := make([]byte, 2)
	_, err = reader.Read(res)
	if err != nil {
		return 0, err
	}
	val = binary.BigEndian.Uint16(res)
	return
}

func WriteUint16(writer io.Writer, val uint16) (err error) {
	res := make([]byte, 2)
	binary.BigEndian.PutUint16(res, val)
	_, err = writer.Write(res)
	return err
}

func ReadInt24(reader io.Reader) (val int32, err error) {
	res := make([]byte, 3)
	_, err = reader.Read(res)
	if err != nil {
		return 0, err
	}
	val = int24.UnmarshalSLE(res)
	return
}

func WriteInt24(writer io.Writer, val int32) (err error) {
	res := int24.MarshalSLE(val)
	_, err = writer.Write(res)
	return err
}

func ReadInt32(reader io.Reader) (val int32, err error) {
	t, err := ReadUint32(reader)
	return int32(t), err
}

func WriteInt32(writer io.Writer, val int32) (err error) {
	return WriteUint32(writer, uint32(val))
}

func ReadUint32(reader io.Reader) (val uint32, err error) {
	res := make([]byte, 4)
	_, err = reader.Read(res)
	if err != nil {
		return 0, err
	}
	val = binary.BigEndian.Uint32(res)
	return
}

func WriteUint32(writer io.Writer, val uint32) (err error) {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, val)
	_, err = writer.Write(res)
	return err
}

func ReadInt64(reader io.Reader) (val int64, err error) {
	t, err := ReadUint64(reader)
	return int64(t), err
}

func WriteInt64(writer io.Writer, val int64) (err error) {
	return WriteUint64(writer, uint64(val))
}

func ReadUint64(reader io.Reader) (val uint64, err error) {
	res := make([]byte, 8)
	_, err = reader.Read(res)
	if err != nil {
		return 0, err
	}
	val = binary.BigEndian.Uint64(res)
	return
}

func WriteUint64(writer io.Writer, val uint64) (err error) {
	res := make([]byte, 8)
	binary.BigEndian.PutUint64(res, val)
	_, err = writer.Write(res)
	return err
}

func ReadUDPAddr(reader io.Reader) (val net.UDPAddr, err error) {
	ipBuf := make([]byte, 5)
	_, err = reader.Read(ipBuf)
	if err != nil {
		return net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 19312,
		}, err
	}
	ip := net.IP(ipBuf[1:])
	port, err := ReadInt16(reader)
	val = net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}
	return
}

func WriteUDPAddr(writer io.Writer, val net.UDPAddr) (err error) {
	err = WriteByte(writer, 4)
	_, err = writer.Write([]byte(val.IP))
	if err != nil {
		return err
	}
	err = WriteUint16(writer, uint16(val.Port))
	return
}

// TODO: Is this ok?
func ReadBoolean(reader io.Reader) (val bool, err error) {
	bb, err := ReadByte(reader)
	if err != nil {
		return false, err
	}
	return bb == 1, nil
}

// TODO: Is this ok?
func WriteBoolean(writer io.Writer, val bool) (err error) {
	v := byte(0)
	if val {
		v = byte(1)
	}
	err = WriteByte(writer, v)
	return
}

func ReadByte(reader io.Reader) (val byte, err error) {
	// If it's really a bytes.Buffer, invoke the direct method for a little more oomph.
	if bb, ok := reader.(*bytes.Buffer); ok {
		val, err = bb.ReadByte()
		return
	}

	// If it's really a bytes.Reader, invoke the direct method for a little more oomph.
	if bb, ok := reader.(*bytes.Reader); ok {
		val, err = bb.ReadByte()
		return
	}

	b := make([]byte, 1)
	_, err = reader.Read(b)
	if err != nil {
		return 0, err
	}
	val = b[0]
	return
}

func WriteByte(writer io.Writer, val byte) (err error) {
	// If it's really a bytes.Buffer, invoke the direct method for a little more oomph.
	if bb, ok := writer.(*bytes.Buffer); ok {
		err = bb.WriteByte(val)
		return
	}

	_, err = writer.Write([]byte{val})
	return
}
