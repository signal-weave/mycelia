package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"mycelia/routing"
	"mycelia/utils"
)

func NewServer(address string, port int) *Server {
	server := &Server{}
	server.router = routing.NewRouter()
	server.address = address
	server.port = port
	return server
}

// Servers are responsible for translating raw TCP string input into routable
// messages.
type Server struct {
	router  *routing.Router
	address string
	port    int
}

// Run ...
func (server *Server) Run() {
	strPort := strconv.Itoa(server.port)
	fullAddress := fmt.Sprintf("%s:%s", server.address, strPort)
	utils.SprintfLn("TCP server on %s", fullAddress)

	listener := utils.ValueOrPanic(net.Listen("tcp", fullAddress))
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err) // Unsure if I want this or to log and continue.
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
		message, err := reader.ReadString('\n')

		if len(message) > 0 {
			go server.router.HandleCommand([]byte(message))
		}

		if err == nil {
			continue
		}

		if err == io.EOF {
			utils.SprintfLn("Client disconnected: %s",
				conn.RemoteAddr().String())
			return // EOF is expected, not an error
		}

		fmt.Println("Error handling message:", err)
		utils.SprintfLn("Client disconnected: %s",
			conn.RemoteAddr().String())
	}
}
