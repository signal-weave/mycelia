package globals

// -----------------------------------------------------------------------------
// Shared, or "global", constants that are referenced between packages.
// This is not meant to contain mutable values.
// -----------------------------------------------------------------------------

const (
	Developer = "Signal Weave"
)

// -------Memory Values---------------------------------------------------------

const (
	BytesInKilobyte = 1024
	BytesInMegabyte = 1024 * BytesInKilobyte
	BytesInGigabyte = 1024 * BytesInMegabyte
)

// The size of the channel partition in size.
const PartitionChanSize = 1024

// -------Verbosity-------------------------------------------------------------

const (
	VERB_NIL = 0 // No printing

	// The VERB_ERR verbosity level is reserved for when internals of the broker
	// encounter a problem that would normally keep the broker from operating.
	VERB_ERR = 1 // Error printing

	// The VERB_WRN verbosity level is reserved for when clients have sent the
	// broker bad data, such as an undialable address, or a fake command.
	VERB_WRN = 2 // Warning + error printing

	// The VERB_ACT verbosity level is reserved for when actions the user may
	// want to know about have taken place, such as when a client connects or
	// disconnects, or when a route as updated.
	VERB_ACT = 3 // Action + warning + error printing
)

// -------Commands--------------------------------------------------------------

const (
	OBJ_UNKNOWN uint8 = 0

	OBJ_DELIVERY    uint8 = 1
	OBJ_TRANSFORMER uint8 = 2
	OBJ_SUBSCRIBER  uint8 = 3

	OBJ_GLOBALS uint8 = 20

	OBJ_Action uint8 = 50
)

const (
	CMD_UNKNOWN uint8 = 0

	CMD_SEND   uint8 = 1
	CMD_ADD    uint8 = 2
	CMD_REMOVE uint8 = 3

	CMD_UPDATE uint8 = 20

	CMD_SIGTERM uint8 = 50
)

// -------Terminal--------------------------------------------------------------

const (
	DEFAULT_TERMINAL_W = 80
	DEFAULT_TERMINAL_H = 25
)
