package cache

import (
	"reflect"
	"slices"
	"testing"
)

// normalizeSerialized turns the SerializeBrokerShape() result (which may have
// nondeterministic route/channel ordering due to map iteration) into a
// deterministic nested map for robust equality checks.
func normalizeSerialized(in *[]map[string]any) map[string]map[string]map[string][]string {
	out := map[string]map[string]map[string][]string{}
	for _, r := range *in {
		routeName, _ := r["name"].(string)
		if _, ok := out[routeName]; !ok {
			out[routeName] = map[string]map[string][]string{}
		}
		channels, _ := r["channels"].([]map[string]any)
		for _, ch := range channels {
			chName, _ := ch["name"].(string)
			tfAny, _ := ch["transformers"].([]map[string]string)
			sbAny, _ := ch["subscribers"].([]map[string]string)

			tf := make([]string, 0, len(tfAny))
			for _, m := range tfAny {
				tf = append(tf, m["address"])
			}
			sb := make([]string, 0, len(sbAny))
			for _, m := range sbAny {
				sb = append(sb, m["address"])
			}

			// Sort for deterministic comparison
			slices.Sort(tf)
			slices.Sort(sb)

			out[routeName][chName] = map[string][]string{
				"transformers": tf,
				"subscribers":  sb,
			}
		}
	}
	return out
}

func TestChannelShape_AddRemove_DeDup(t *testing.T) {
	// Fresh global for test isolation
	BrokerShape = NewBrokerShape()

	cs := BrokerShape.Route("r").Channel("c")

	// Add transformers (with duplicate)
	cs.AddTransformer("127.0.0.1:7010")
	cs.AddTransformer("10.0.0.52:8008")
	cs.AddTransformer("127.0.0.1:7010") // dup ignored

	if got := len(cs.Transformers); got != 2 {
		t.Fatalf("expected 2 transformers, got %d", got)
	}

	// Remove one transformer
	cs.RemoveTransformer("127.0.0.1:7010")
	if got := len(cs.Transformers); got != 1 {
		t.Fatalf("expected 1 transformer after remove, got %d", got)
	}
	if cs.Transformers[0] != "10.0.0.52:8008" {
		t.Fatalf("unexpected remaining transformer: %v", cs.Transformers)
	}

	// Add subscribers (with duplicate)
	cs.AddSubscriber("127.0.0.1:1234")
	cs.AddSubscriber("16.70.18.1:9999")
	cs.AddSubscriber("127.0.0.1:1234") // dup ignored

	if got := len(cs.Subscribers); got != 2 {
		t.Fatalf("expected 2 subscribers, got %d", got)
	}

	// Remove subscriber
	cs.RemoveSubscriber("16.70.18.1:9999")
	if got := len(cs.Subscribers); got != 1 {
		t.Fatalf("expected 1 subscriber after remove, got %d", got)
	}
	if cs.Subscribers[0] != "127.0.0.1:1234" {
		t.Fatalf("unexpected remaining subscriber: %v", cs.Subscribers)
	}
}

func TestSerializeBrokerShape_Empty(t *testing.T) {
	BrokerShape = NewBrokerShape()

	ser, err := SerializeBrokerShape()
	if err != nil {
		t.Fatalf("SerializeBrokerShape error: %v", err)
	}
	if ser == nil {
		t.Fatal("SerializeBrokerShape returned nil slice pointer")
	}
	if len(*ser) != 0 {
		t.Fatalf("expected 0 routes, got %d", len(*ser))
	}
}

func TestSerializeBrokerShape_Populated(t *testing.T) {
	BrokerShape = NewBrokerShape()

	// Build: routes/default/channels/inmem with two transformers + two subscribers
	ch := BrokerShape.Route("default").Channel("inmem")
	ch.AddTransformer("127.0.0.1:7010")
	ch.AddTransformer("10.0.0.52:8008")
	ch.AddSubscriber("127.0.0.1:1234")
	ch.AddSubscriber("16.70.18.1:9999")

	ser, err := SerializeBrokerShape()
	if err != nil {
		t.Fatalf("SerializeBrokerShape error: %v", err)
	}
	if ser == nil {
		t.Fatal("SerializeBrokerShape returned nil slice pointer")
	}

	got := normalizeSerialized(ser)
	want := map[string]map[string]map[string][]string{
		"default": {
			"inmem": {
				"transformers": []string{"10.0.0.52:8008", "127.0.0.1:7010"},
				"subscribers":  []string{"127.0.0.1:1234", "16.70.18.1:9999"},
			},
		},
	}
	tf := want["default"]["inmem"]["transformers"]
	sb := want["default"]["inmem"]["subscribers"]
	slices.Sort(tf)
	slices.Sort(sb)
	want["default"]["inmem"]["transformers"] = tf
	want["default"]["inmem"]["subscribers"] = sb

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("serialized structure mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}
