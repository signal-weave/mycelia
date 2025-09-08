package shutdown

import (
	"mycelia/logging"
)

func Shutdown() {
	logging.LogSystemAction("Starting shutdown Process!")

	logging.GlobalLogger.Stop()
}
