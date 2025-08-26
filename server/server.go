package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"mycelia/errgo"
	"mycelia/routing"
	"mycelia/str"
)

func NewServer(address string, port int) *Server {
	server := &Server{}
	server.Broker = routing.NewBroker()
	server.address = address
	server.port = port
	return server
}

// Servers are responsible for translating raw TCP string input into routable
// messages.
type Server struct {
	Broker  *routing.Broker
	address string
	port    int
}

// Run ...
func (server *Server) Run() {
	strPort := strconv.Itoa(server.port)
	fullAddress := fmt.Sprintf("%s:%s", server.address, strPort)
	str.SprintfLn("TCP server on %s", fullAddress)

	listener := errgo.ValueOrPanic(net.Listen("tcp", fullAddress))
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			// TODO: Print values for address of sender.
			str.ErrorPrint("Listener could not accept message.")
			continue
		}
		go server.handleConnection(conn)
	}
}

// Handle incoming data stream.
func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	aMsg := fmt.Sprintf("Client connected: %s\n", conn.RemoteAddr().String())
	str.ActionPrint(aMsg)

	for {
		frame, err := readFrame(conn)
		if err != nil {
			// EOF or other error—close connection
			return
		}
		if len(frame) == 0 {
			continue
		}

		server.Broker.HandleBytes(frame)
	}
}

// Read the frame's byte stream until the message header's worth of bytes have
// been consumed, then return a buffer of those bytes or error.
func readFrame(conn net.Conn) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(conn, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	if n == 0 {
		return []byte{}, nil
	}

	buf := make([]byte, n)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
