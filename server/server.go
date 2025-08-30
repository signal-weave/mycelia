package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"

	"mycelia/globals"
	"mycelia/routing"
	"mycelia/str"
)

func NewServer(address string, port int) *Server {
	server := &Server{}
	server.Broker = routing.NewBroker()
	server.Broker.ManagingServer = server
	server.address = address
	server.port = port
	return server
}

// Servers are responsible for translating raw TCP string input into routable
// messages.
type Server struct {
	Broker   *routing.Broker
	address  string
	port     int
	listener net.Listener
	mutex    sync.RWMutex
}

func (s *Server) GetAddress() string {
	return s.address
}

func (s *Server) GetPort() int {
	return s.port
}

// Run ...
func (server *Server) Run() {
	if server.listener == nil {
		server.UpdateListener()
	}

	strPort := strconv.Itoa(server.port)
	fullAddress := fmt.Sprintf("%s:%s", server.address, strPort)
	str.SprintfLn("Listening on %s", fullAddress)

	// TODO: Print values for address of sender.
	for {
		server.mutex.RLock()
		l := server.listener
		server.mutex.RUnlock()

		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				continue
			}
			eMsg := fmt.Sprintf("Listener accept error %v", err)
			str.ErrorPrint(eMsg)
			continue
		}
		go server.HandleConnection(conn)
	}
}

// Updates the socket the server is listening to at runtime.
func (server *Server) UpdateListener() {
	// open new first
	addr := fmt.Sprintf("%s:%d", globals.Address, globals.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not open listener to %s, staying on old listener!", addr,
		)
		str.WarningPrint(wMsg)

		// These should be in sync, not thrilled with this here though.
		globals.Address = server.address
		globals.Port = server.port
		return
	}

	server.mutex.Lock()
	old := server.listener
	server.listener = l
	server.address = globals.Address
	server.port = globals.Port
	server.mutex.Unlock()

	str.SprintfLn("Now listening on %s", addr)

	if old != nil {
		str.SprintfLn("Closing listener on %s", old.Addr().String())
		_ = old.Close() // will cause a benign net.ErrClosed in Run()
	}
}

// Handle incoming data stream.
func (server *Server) HandleConnection(conn net.Conn) {
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
	const lenU32 = 4
	var hdr [lenU32]byte
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
