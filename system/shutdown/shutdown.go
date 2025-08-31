package shutdown

import (
	"mycelia/system/cache"
)

// All steps necessary to perform a shutdown procedure.
func PerformShutdown() {
	writeShutdownReport()
}

// Snapshot the shape of the broker + runtime parameters and write it out as a
// shutdown report and denotes a graceful shutdown.
func writeShutdownReport() {
	cache.WriteReport(true)
}
