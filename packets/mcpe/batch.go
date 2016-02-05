package mcpe

import (
	"../raknet"
	"bytes"
	"compress/zlib"
	"io"
)

type MCPEBatch struct {
	Payload [][]byte
}

// Useful helper function to add encodable packets for the payload.
func (pkt *MCPEBatch) AddPacket(p raknet.EncodablePacket) (err error) {
	buf := new(bytes.Buffer)
	err = p.Encode(buf)
	if err != nil {
		return
	}

	pkt.Payload = append(pkt.Payload, buf.Bytes())
	return
}

func (pkt *MCPEBatch) Id() byte {
	return ID_MCPE_BATCH
}

func (pkt *MCPEBatch) Decode(reader io.Reader) (err error) {
	// FIXME: Do we need to bother?
	_, err = raknet.ReadInt32(reader)

	r, err := zlib.NewReader(reader)
	if err != nil {
		return
	}
	defer r.Close()

	var pkts [][]byte
	for {
		ln, err := raknet.ReadInt32(r)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		data := make([]byte, ln)
		_, err = r.Read(data)
		if err != nil {
			return err
		}

		pkts = append(pkts, data)
	}

	pkt.Payload = pkts
	return
}

func (pkt MCPEBatch) Encode(writer io.Writer) (err error) {
	err = raknet.WriteByte(writer, ID_MCPE_BATCH)
	if err != nil {
		return
	}

	// The compression has to be buffered as we must write a length.
	buffer := new(bytes.Buffer)
	w := zlib.NewWriter(buffer)

	for _, payload := range pkt.Payload {
		if err = raknet.WriteInt32(w, int32(len(payload))); err != nil {
			return err
		}

		if _, err = w.Write(payload); err != nil {
			return
		}
	}

	if err = w.Close(); err != nil {
		return
	}

	err = raknet.WriteInt32(writer, int32(buffer.Len()))
	_, err = io.Copy(writer, buffer)
	return
}
