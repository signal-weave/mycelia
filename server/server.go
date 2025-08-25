package server

import (
	"bufio"
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
	fmt.Printf("Client connected: %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	for {
		msgLen, err := binary.ReadUvarint(reader)

		if err != nil {
			if err == io.EOF {
				str.SprintfLn(
					"Client disconnected: %s", conn.RemoteAddr().String(),
				)
				return
			}
			// Any other error => connection/read framing error
			str.WarningPrint(fmt.Sprintf("Bad message length: %v", err))
			return
		}

		if msgLen == 0 {
			// empty message
			continue
		}

		msg := make([]byte, msgLen)
		if _, err := io.ReadFull(reader, msg); err != nil {
			if err == io.EOF {
				// EOF is expected, not an error
				str.SprintfLn(
					"Client disconnected: %s", conn.RemoteAddr().String(),
				)
				return
			}
			str.WarningPrint(fmt.Sprintf("Bad message body: %v", err))
			return
		}

		go server.Broker.HandleBytes(msg)
	}
}
