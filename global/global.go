package global

// -----------------------------------------------------------------------------
// Shared, or "global", constants that are referenced between packages.
// This is not meant to contain mutable values.
// -----------------------------------------------------------------------------

// -------Commands--------------------------------------------------------------

const (
	OBJ_DELIVERY    uint8 = 1
	OBJ_TRANSFORMER uint8 = 2
	OBJ_SUBSCRIBER  uint8 = 3
)

const (
	CMD_UNKNOWN uint8 = iota
	CMD_SEND    uint8 = 1
	CMD_ADD     uint8 = 2
	CMD_REMOVE  uint8 = 3
)
