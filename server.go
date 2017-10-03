package main

// Server interface for any relaybroker server
type Server interface {
	Start()
}

type tcpServer struct {
	port int
}

// NewServer create server
func NewServer(port int) Server {
	return tcpServer{
		port: port,
	}
}

func (s tcpServer) Start() {

}
