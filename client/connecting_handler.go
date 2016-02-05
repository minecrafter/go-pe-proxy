package client

import (
	"../packets/mcpe"
	"../packets/raknet"
	"../player"
	"bytes"
	"math/rand"
	"time"
)

type BackendConnectingHandler struct {
	backendConnection proxy.BackendConnection
	player            *player.Connection
}

func (h *BackendConnectingHandler) OnMessage(buf []byte) {
	pktData := buf[1:]

	switch buf[0] {
	case raknet.ID_CONNECTED_PONG:
		pkt := new(raknet.RakNetConnectedPong)
		err := pkt.Decode(bytes.NewReader(pktData))
		if err != nil {
			log.Printf("Error whilst handling message from %s", endpoint.String(), err)
			return
		}

		log.Printf("Handling a connected ping packet from %s.", endpoint.String())
		bc.TouchUpdated()
		break
	case raknet.ID_DATA_4:
		// TODO: Should implement states.

		// Datagram with encapsulated packet(s), not split.
		pkt := new(raknet.RakNetDatagram)
		err := pkt.Decode(buf)
		if err != nil {
			log.Printf("Error whilst handling message from %s", endpoint.String(), err)
			return
		}

		for _, item := range pkt.Payload {
			log.Printf("Handling extracted packet: %s ...", hex.EncodeToString(item.Payload[0:15]))
			bc.handleIncoming(item.Payload, endpoint)
		}
		break
	case raknet.ID_OPEN_CONNECTION_REPLY_1:
		pkt := new(raknet.RakNetOpenConnectionReply1)
		err := pkt.Decode(bytes.NewReader(pktData))
		if err != nil {
			log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
			bc.player.connectionFailure(bc.server, err)
			return
		}

		// Save the MTU, and send a second packet
		bc.mtu = pkt.MTU
		next := raknet.RakNetOpenConnectionRequest2{
			ClientEndpoint: bc.connection.LocalAddr().(*net.UDPAddr),
			MTU:            bc.mtu,
			GUID:           bc.guid,
		}
		if err := bc.SendPacket(next); err != nil {
			log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
			bc.player.connectionFailure(bc.server, err)
			return
		}
		break
	case raknet.ID_OPEN_CONNECTION_REPLY_2:
		pkt := new(raknet.RakNetOpenConnectionReply2)
		err := pkt.Decode(bytes.NewReader(pktData))
		if err != nil {
			log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
			bc.player.connectionFailure(bc.server, err)
			return
		}

		// We ignore the decoded packet, its contents are useless.
		// Send the Connection Request packet.
		next := raknet.RakNetConnectionRequest{
			GUID:      bc.guid,
			Timestamp: time.Now().UnixNano(),
			Security:  0,
		}

		// Need to encapsulate it in a datagram
		datagrams, err := raknet.CreateDatagrams(bc.reliabilityNumber,
			[]raknet.EncodablePacket{next})
		if err != nil {
			log.Printf("Error whilst handling message from %s", endpoint.String(), err)
			return
		}

		for _, datagram := range *datagrams {
			if err := bc.SendPacket(datagram); err != nil {
				log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
				bc.player.connectionFailure(bc.server, err)
				return
			}
		}
		break
	case raknet.ID_CONNECTION_REQUEST_ACCEPTED:
		pkt := new(raknet.RakNetConnectionRequestAccepted)
		err := pkt.Decode(bytes.NewReader(pktData))
		if err != nil {
			log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
			bc.player.connectionFailure(bc.server, err)
			return
		}

		// Ignore this packet too. Send NEW_INCOMING_CONNECTION now.
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		next := raknet.RakNetNewIncomingConnection{
			Cookie:   r.Int31(),
			Secure:   0,
			Port:     int16((&bc.server.Address).Port),
			Session1: r.Int63(),
			Session2: r.Int63(),
		}
		if err := bc.SendPacket(next); err != nil {
			log.Printf("Error whilst handling server message from %s", endpoint.String(), err)
			bc.player.connectionFailure(bc.server, err)
			return
		}

		break
	}
}

func (h *BackendConnectingHandler) OnClose() {

}
