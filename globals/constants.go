package globals

// -----------------------------------------------------------------------------
// Shared, or "global", constants that are referenced between packages.
// This is not meant to contain mutable values.
// -----------------------------------------------------------------------------

// -------Misc------------------------------------------------------------------

const (
	Developer = "Signal Weave"
)

const (
	DateLayout = "01-02-2006"  // MM-DD-YYYY
	TimeLayout = "15-04-05-00" // HH-MM-SS-XX
)

// -------Memory Values---------------------------------------------------------

const (
	BytesInKilobyte = 1024
	BytesInMegabyte = 1024 * BytesInKilobyte
	BytesInGigabyte = 1024 * BytesInMegabyte
)

// -------Verbosity-------------------------------------------------------------

const (
	VERB_NIL = 0 // No printing

	// The VERB_ERR verbosity level is reserved for when internals of the broker
	// encounter a problem that would normally keep the broker from operating.
	VERB_ERR = 1 // Error printing

	// The VERB_WRN verbosity level is reserved for when clients have sent the
	// broker bad data, such as an undialable address, or a fake object.
	VERB_WRN = 2 // Warning + error printing

	// The VERB_ACT verbosity level is reserved for when actions the user may
	// want to know about have taken place, such as when a client connects or
	// disconnects, or when a route as updated.
	VERB_ACT = 3 // Action + warning + error printing
)

// -------Objects---------------------------------------------------------------

const (
	OBJ_UNKNOWN uint8 = 0

	OBJ_DELIVERY    uint8 = 1
	OBJ_TRANSFORMER uint8 = 2
	OBJ_SUBSCRIBER  uint8 = 3
	OBJ_CHANNEL     uint8 = 4

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

// -------Acks/Nacks------------------------------------------------------------

const (
	// Sender does not wish to receive ack.
	ACK_PLCY_NOREPLY uint8 = 0

	// Sender wants to get ack when broker delivers to final subscriber.
	// This often means sending the ack back after the final channel has
	// processed the message object.
	ACK_PLCY_ONSENT uint8 = 1
)

const (
	ACK_UNKNOWN uint8 = 0 // Undetermined

	// Broker was able to and finished sending message to subscribers.
	ACK_SENT uint8 = 1

	// This isn't used by the broker, but its here for clarity.
	// Client APIs do use this value when timing out while trying to connect to
	// the broker.
	// If no ack was gotten before the timeout time, a response with ACK_TIMEOUT
	// is generated and returned instead.
	ACK_TIMEOUT uint8 = 10

	ACK_CHANNEL_NOT_FOUND      uint8 = 20
	ACK_CHANNEL_ALREADY_EXISTS uint8 = 21
	ACK_ROUTE_NOT_FOUND        uint8 = 30
)

// -------Terminal--------------------------------------------------------------

const (
	DEFAULT_TERMINAL_W = 80
	DEFAULT_TERMINAL_H = 25
)
