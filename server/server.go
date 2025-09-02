package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"mycelia/comm"
	"mycelia/globals"
	"mycelia/routing"
	"mycelia/str"
)

func NewServer(address string, port int) *Server {
	server := &Server{}
	server.Broker = routing.NewBroker(server)
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

	for !globals.PerformShutdown.Load() {
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

func (server *Server) Shutdown() {
	globals.PerformShutdown.Store(true)

	server.mutex.Lock()
	l := server.listener
	server.listener = nil
	server.mutex.Unlock()

	if l != nil {
		_ = l.Close()
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

	str.ActionPrint(
		fmt.Sprintf("Processing message from %s", conn.RemoteAddr().String()),
	)

	for {
		frame, err := comm.ReadFrameU32(conn)
		if err != nil {
			_ = comm.WriteFrameU32(
				conn, []byte("ERR: invalid frame:"+err.Error()),
			)
			return
		}
		if len(frame) == 0 {
			continue
		}

		server.Broker.HandleBytes(frame)
	}
}
