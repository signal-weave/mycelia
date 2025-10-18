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

	"github.com/signal-weave/rhizome"
)

func NewServer(address string, port int) *Server {
	server := &Server{}
	server.Broker = routing.NewBroker(server)
	server.address = address
	server.port = port
	return server
}

// Server is responsible for translating raw TCP string input into routable
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
func (s *Server) Run() error {
	if s.listener == nil {
		if err := s.UpdateListener(); err != nil {
			return err
		}
	}
	return s.serve()
}

// Spins up the server...
// Caps concurrency and avoids spawning unbound go routines.
func (s *Server) serve() error {
	if s.jobs == nil {
		s.jobs = make(chan net.Conn, 1024)
	}

	for range globals.WorkerCount {
		go func() {
			for c := range s.jobs {
				s.HandleConnection(c)
			}
		}()
	}

	for !globals.PerformShutdown.Load() {
		c, err := s.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		// Go runtime selects an unblocked worker when we push the
		// listener.Accept() into a channel of multiple objects.
		s.jobs <- c
	}

	return nil
}

func (s *Server) Shutdown() {
	globals.PerformShutdown.Store(true)

	s.mutex.Lock()
	l := s.listener
	s.listener = nil
	s.mutex.Unlock()

	if l != nil {
		_ = l.Close()
	}
	if s.jobs != nil {
		close(s.jobs)
	}
}

// UpdateListener updates which socket the server is listening to at runtime.
func (s *Server) UpdateListener() error {
	// open new first
	addr := fmt.Sprintf("%s:%d", globals.Address, globals.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		wMsg := fmt.Sprintf(
			"Could not open listener to %s, staying on old listener!", addr,
		)
		logging.LogSystemError(wMsg)

		// These should be in sync, not thrilled with this here though.
		globals.Address = s.address
		globals.Port = s.port
		return err
	}

	s.mutex.Lock()
	old := s.listener
	s.listener = l
	s.address = globals.Address
	s.port = globals.Port
	s.mutex.Unlock()

	str.SprintfLn("Now listening on %s", addr)

	if old != nil {
		str.SprintfLn("Closing listener on %s", old.Addr().String())
		_ = old.Close() // will cause a benign net.ErrClosed in Run()
	}

	return nil
}

// HandleConnection manages incoming data stream.
func (s *Server) HandleConnection(conn net.Conn) {
	logging.LogSystemAction(
		fmt.Sprintf("Client connected: %s\n", conn.RemoteAddr().String()),
	)

	resp := rhizome.NewConnResponder(conn)

	for {
		frame, err := comm.ReadFrameU32(conn)
		if err != nil {
			return
		}
		if len(frame) == 0 {
			continue
		}

		s.Broker.HandleBytes(frame, resp)
	}
}
