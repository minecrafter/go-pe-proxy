package proxy

import (
	"log"
	"math/rand"
	"net"
	"runtime"
)

type Proxy struct {
	Registry *SessionRegistry

	address *net.UDPAddr
	conn    *net.UDPConn
	guid    int64

	unknownSession *unknownSession
	servers        map[string]*Server
}

func NewProxy(address *net.UDPAddr) (this *Proxy) {
	this = new(Proxy)
	this.Registry = NewSessionRegistry()
	this.address = address
	this.unknownSession = NewUnknownSession(this)
	this.servers = make(map[string]*Server)
	this.guid = rand.Int63()
	return
}

func (this *Proxy) Close() {
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
}

func (this *Proxy) ListenAndServe() {
	conn, err := net.ListenUDP("udp", this.address)
	if err != nil {
		log.Printf("Unable to bind to %s: %s", this.address.String(), err.Error())
		return
	}

	// Start some goroutines to handle "unknown session" packets
	for i := 0; i < runtime.NumCPU(); i++ {
		go this.unknownSession.Process()
	}

	this.conn = conn

	// Begin serving clients in perpetuity.
	for {
		buf := make([]byte, 1500)
		read, endpoint, err := conn.ReadFromUDP(buf)

		if err != nil {
			log.Println("Encountered an error while listening:", err)
			return
		}

		// TODO: The performance of this will suck big-time. Find a more efficient
		// replacement!
		conn := this.Registry.GetByEndpoint(endpoint)

		if conn != nil {
			conn.processQueue <- buf[0:read]
		} else {
			entry := unknownSessionEntry{buf[0:read], endpoint}
			this.unknownSession.processQueue <- entry
		}
	}
}
