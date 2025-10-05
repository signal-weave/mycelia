package shutdown

import (
	"mycelia/logging"

	"github.com/signal-weave/siglog"
)

func Shutdown() {
	logging.LogSystemAction("Starting shutdown Process!")

	siglog.Shutdown()
}
