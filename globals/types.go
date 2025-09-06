package globals

// -------Channels--------------------------------------------------------------

type SelectionStrategy int

const (
	// Default name for deadletter channels.
	DeadLetter = "deadLetter"

	// The number of protocol.Object that a mycelia channel partition can hold
	// at any maximum.
	PartitionChanSize = 128
)

const (
	SelStratRandom SelectionStrategy = iota
	SelStratRoundRobin
	SelStratPubSub
)

var StrategyName = map[SelectionStrategy]string{
	SelStratRandom:     "random",
	SelStratRoundRobin: "round-robin",
	SelStratPubSub:     "pub-sub",
}

var StrategyValue = map[string]SelectionStrategy{
	"random":      SelStratRandom,
	"round-robin": SelStratRoundRobin,
	"pub-sub":     SelStratPubSub,
}

func (ss SelectionStrategy) String() string {
	return StrategyName[ss]
}
