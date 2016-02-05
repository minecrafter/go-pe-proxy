package proxy

import "net"

type Server struct {
  Name string
  Address *net.UDPAddr
}

func NewServer(name string, address *net.UDPAddr) (this *Server) {
  this = new(Server)
  this.Name = name
  this.Address = address
  return this
}
