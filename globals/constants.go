package globals

// -----------------------------------------------------------------------------
// Shared, or "global", constants that are referenced between packages.
// This is not meant to contain mutable values.
// -----------------------------------------------------------------------------

// -------Misc------------------------------------------------------------------

const (
	Developer = "Signal Weave"
)

// -------Memory Values---------------------------------------------------------

const (
	BytesInKilobyte = 1024
	BytesInMegabyte = 1024 * BytesInKilobyte
)

// -------Logging---------------------------------------------------------------

const (
	VerbNil = 0 // No printing

	// The VerbErr verbosity level is reserved for when internals of the broker
	// encounter a problem that would normally keep the broker from operating.
	VerbErr = 1 // Error printing

	// The VerbWrn verbosity level is reserved for when clients have sent the
	// broker bad data, such as an undialable address, or a fake object.
	VerbWrn = 2 // Warning + error printing

	// The VertAct verbosity level is reserved for when actions the user may
	// want to know about have taken place, such as when a client connects or
	// disconnects, or when a route as updated.
	VertAct = 3 // Action + warning + error printing
)

const (
	LogToFile    = 0
	LogToConsole = 1
	LogToBoth    = 2
)

// -------Objects---------------------------------------------------------------

const (
	ObjUnknown uint8 = 0

	ObjDelivery    uint8 = 1
	ObjTransformer uint8 = 2
	ObjSubscriber  uint8 = 3
	ObjChannel     uint8 = 4

	ObjGlobals uint8 = 20

	ObjAction uint8 = 50
)

const (
	CmdUnknown uint8 = 0

	CmdSend   uint8 = 1
	CmdAdd    uint8 = 2
	CmdRemove uint8 = 3

	CmdUpdate uint8 = 20

	CmdSigterm uint8 = 50
)

// -------Acks/Nacks------------------------------------------------------------

const (
	// AckPlcyNoreply means sender does not wish to receive ack.
	AckPlcyNoreply uint8 = 0

	// AckPlcyOnsent means sender wants to get ack when broker delivers to final
	//subscriber.
	// This often means sending the ack back after the final channel has
	// processed the message object.
	AckPlcyOnsent uint8 = 1
)

const (
	AckUnknown uint8 = 0 // Undetermined

	// AckSent means broker was able to and finished sending message to
	//subscribers.
	AckSent uint8 = 1

	// AckTimeout isn't used by the broker, but its here for clarity.
	// Client APIs do use this value when timing out while trying to connect to
	// the broker.
	// If no ack was gotten before the timeout time, a response with AckTimeout
	// is generated and returned instead.
	AckTimeout uint8 = 10

	AckChannelNotFound      uint8 = 20
	AckChannelAlreadyExists uint8 = 21
	AckRouteNotFound        uint8 = 30
)

// -------Terminal--------------------------------------------------------------

const (
	DefaultTerminalW = 80
	DefaultTerminalH = 25
)
