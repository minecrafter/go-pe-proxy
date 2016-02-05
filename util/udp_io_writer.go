package util

import "net"

type UdpIoWriter struct {
	conn     *net.UDPConn
	endpoint *net.UDPAddr
}

func NewUdpIoWriter(conn *net.UDPConn, endpoint *net.UDPAddr) UdpIoWriter {
	return UdpIoWriter{conn, endpoint}
}

func (w UdpIoWriter) Write(p []byte) (n int, err error) {
	n, err = w.conn.WriteToUDP(p, w.endpoint)
	return
}
