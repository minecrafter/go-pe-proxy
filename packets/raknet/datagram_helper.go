package raknet

import (
	"log"
	"sync"
	"time"
)

type DatagramSender interface {
	IsAlive() bool
	GetEndpointString() string
	SendPacket(data EncodablePacket) error
}

type sentDatagram struct {
	data  RakNetDatagram
	sent  time.Time
	tries int
}

type DatagramHelper struct {
	sentDatagrams map[int32]*sentDatagram
	toSend        DatagramSender
	sync.Mutex
}

func NewDatagramHelper(toSend DatagramSender) (this *DatagramHelper) {
	this = new(DatagramHelper)
	this.sentDatagrams = make(map[int32]*sentDatagram)
	this.toSend = toSend
	return
}

func (this *DatagramHelper) RegisterDatagram(datagram RakNetDatagram) {
	this.Lock()
	if _, ok := this.sentDatagrams[datagram.DatagramSequenceNumber]; ok {
		log.Printf("Tried to register already known datagram %d!", datagram.DatagramSequenceNumber)
	} else {
		this.sentDatagrams[datagram.DatagramSequenceNumber] = &sentDatagram{
			data:  datagram,
			sent:  time.Now(),
			tries: 0,
		}
	}
	this.Unlock()
}

func (this *DatagramHelper) HandleAck(ack *RakNetAck) {
	this.Lock()
	for _, item := range ack.Acknowledged {
		for id := item.Min; id <= item.Max; id++ {
			as32 := int32(id)
			if _, ok := this.sentDatagrams[as32]; ok {
				log.Printf("Marked %d as ACK.", id)
				delete(this.sentDatagrams, as32)
			} else {
				log.Printf("Tried to ack unknown datagram %d!", id)
			}
		}
	}
	this.Unlock()
}

func (this *DatagramHelper) HandleNak(nak *RakNetNak) {
	this.Lock()
	for _, item := range nak.NotAcknowledged {
		for id := item.Min; id <= item.Max; id++ {
			as32 := int32(id)
			if _, ok := this.sentDatagrams[as32]; ok {
				log.Printf("Marked %d as NAK!", id)
				delete(this.sentDatagrams, as32)
			} else {
				log.Printf("Tried to nak unknown datagram %d!", id)
			}
		}
	}
	this.Unlock()
}

func fib(n int) int {
	prev := 0
	next := 1
	for i := 0; i < n; i++ {
		tmp := next
		next = next + prev
		prev = tmp
	}
	return next
}

func (this *DatagramHelper) TryResendPackets() {
	now := time.Now()
	this.Lock()
	for k, v := range this.sentDatagrams {
		// TODO: More fancy datagram resend code would be nice, but I can't be
		// arsed to implement that right now.
		delay := time.Duration(fib(12+v.tries)) * time.Millisecond
		if v.sent.Add(delay).Before(now) {
			if v.tries >= 2 {
				log.Printf("Datagram %d lost for %s.", k, this.toSend.GetEndpointString())
				delete(this.sentDatagrams, k)
				continue
			}
			// Try resending this packet
			if err := this.toSend.SendPacket(v.data); err != nil {
				log.Printf("Datagram %d lost for %s due to error: %s", k,
					this.toSend.GetEndpointString(), err.Error())
				delete(this.sentDatagrams, k)
				continue
			}

			v.sent = now
			v.tries = v.tries + 1
		}
	}
	this.Unlock()
}
