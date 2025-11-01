package globals

// -------Channels--------------------------------------------------------------

// SelectionStrategy is the enum for which selection strategy a channel
// utilizes. It is placed in globals so that system/startup can reference it
// when parsing Mycelia_Config.json files, without needing to import the routing
// package.
type SelectionStrategy int

const (
	// DeadLetter is the name for dead letter channels.
	DeadLetter = "deadLetter"

	// PartitionChanSize is the maximum number of protocol.Object that a mycelia
	// channel partition can hold at any time.
	PartitionChanSize = 128
)

const (
	SelStratRandom SelectionStrategy = 1 << iota
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
