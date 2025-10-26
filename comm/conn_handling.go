package comm

import (
	"fmt"
	"net"

	"mycelia/logging"
)

// CloseConnection simply closes the conn and logs any possible error.
func CloseConnection(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		m := fmt.Sprintf(
			"Unable to close connection to %s: %s", conn.RemoteAddr(), err,
		)
		logging.LogSystemWarning(m)
	}
}
