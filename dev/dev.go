package dev

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// -----------------------------------------------------------------------------
// A library of various debugging and development utilities.
// -----------------------------------------------------------------------------

// BlockUntilInterrupt holds the program termination until ctrl + c is entered.
func BlockUntilInterrupt() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs

	fmt.Println("\nCtrl+C detected. Exiting gracefully.")
}
