package routing

import (
	"testing"
	"time"

	"mycelia/globals"
	"mycelia/protocol"
	"mycelia/test"
)

// dummyServer implements the minimal broker.ManagingServer interface.
// It avoids circular deps with your real server type.
type dummyServer struct {
	addr string
	port int
}

func (d *dummyServer) UpdateListener()    {}
func (d *dummyServer) GetAddress() string { return d.addr }
func (d *dummyServer) GetPort() int       { return d.port }
func (d *dummyServer) Shutdown()          {}

func TestBroker_DeliveryThroughTransformerToSubscriber(t *testing.T) {
	// --- Globals tuned for the test run
	globals.DefaultNumPartitions = 1
	globals.TransformTimeout = 1 * time.Second
	globals.AutoConsolidate = false
	globals.PrintTree = false

	// --- Mocks: one transformer (echo prefix) + one subscriber (capture body)
	xformPrefix := "X:"
	xformAddr, stopX := test.MockTwoWayServer(t, xformPrefix)
	defer stopX()

	subAddr, gotBody, stopSub := test.MockOneWayServer(t)
	defer stopSub()

	// --- Broker under test
	br := NewBroker(&dummyServer{addr: "127.0.0.1", port: 0})

	const route = "default"
	const channel = "inmem"

	// 1) Add transformer
	addX := protocol.NewObject(
		globals.OBJ_TRANSFORMER,
		globals.CMD_ADD,
		"tester", "uid-x",
		route, channel, xformAddr, "",
		nil,
	)
	if err := br.HandleObject(addX); err != nil {
		t.Fatalf("add transformer failed: %v", err)
	}

	// 2) Add subscriber
	addS := protocol.NewObject(
		globals.OBJ_SUBSCRIBER,
		globals.CMD_ADD,
		"tester", "uid-s",
		route, channel, subAddr, "",
		nil,
	)
	if err := br.HandleObject(addS); err != nil {
		t.Fatalf("add subscriber failed: %v", err)
	}

	// 3) Send a delivery
	orig := "hello!"
	send := protocol.NewObject(
		globals.OBJ_DELIVERY,
		globals.CMD_SEND,
		"client-1", "uid-d",
		route, channel, "hash-key", "",
		[]byte(orig),
	)
	if err := br.HandleObject(send); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	// 4) The subscriber should receive the transformed payload
	select {
	case got := <-gotBody:
		want := xformPrefix + orig
		if got != want {
			t.Fatalf("subscriber got %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for subscriber to receive message")
	}
}
