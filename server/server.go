package server

import (
	"fmt"
	"net"
	"sync"

	"mycelia/comm"
	"mycelia/globals"
	"mycelia/logging"
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
	jobs     chan net.Conn
	listener net.Listener

	Broker *routing.Broker

	address string
	port    int

	mutex sync.RWMutex
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
	server.serve()
}

// Spins up the server...
// Caps concurrency and avoids spawning unbound go routines.
func (server *Server) serve() error {
	if server.jobs == nil {
		server.jobs = make(chan net.Conn, 1024)
	}

	for range globals.WorkerCount {
		go func() {
			for c := range server.jobs {
				server.HandleConnection(c)
			}
		}()
	}

	for !globals.PerformShutdown.Load() {
		c, err := server.listener.Accept()
		if err != nil {
			return err
		}

		// Go runtime selects an unblocked worker when we push the
		// listener.Accept() into a channel of multiple objects.
		server.jobs <- c
	}

	return nil
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
	if server.jobs != nil {
		close(server.jobs)
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
		logging.LogSystemError(wMsg)

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
	logging.LogSystemAction(
		fmt.Sprintf("Client connected: %s\n", conn.RemoteAddr().String()),
	)

	resp := comm.NewConnResponder(conn)

	for {
		frame, err := comm.ReadFrameU32(conn)
		if err != nil {
			return
		}
		if len(frame) == 0 {
			continue
		}

		server.Broker.HandleBytes(frame, resp)
	}
}
