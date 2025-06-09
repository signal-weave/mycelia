package commands

// Status of the current message describing its life cycle or validity.
type Status int

const (
	// The initial status of a valid message.
	StatusCreated = iota
	// When a message is waiting in queue to be sent to next consumer.
	StatusPending
	// Message has been consumed.
	StatusResolved
	// A message whose source string could not be properly decoded.
	StatusInvalid
)
