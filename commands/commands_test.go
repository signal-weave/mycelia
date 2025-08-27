package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestStatusConstants(t *testing.T) {
	if StatusCreated != 0 {
		t.Errorf("StatusCreated expected 0, got %d", StatusCreated)
	}
	if StatusResolved != 1 {
		t.Errorf("StatusResolved expected 2, got %d", StatusResolved)
	}
	if StatusInvalid != 2 {
		t.Errorf("StatusInvalid expected 3, got %d", StatusInvalid)
	}
}

func TestMessageStruct(t *testing.T) {
	msg := Message{
		Cmd:    uint8(1),
		ID:     "123",
		Route:  "main",
		Status: StatusCreated,
		Body:   []byte("Hello!"),
	}

	if msg.Cmd != uint8(1) || msg.ID != "123" || msg.Route != "main" ||
		msg.Status != StatusCreated || !bytes.Equal(msg.Body, []byte("Hello!")) {
		t.Errorf("Message fields not assigned correctly: %+v", msg)
	}
}

func TestSubscriberStruct(t *testing.T) {
	sub := Subscriber{
		Cmd:     uint8(2),
		ID:      "sub-1",
		Route:   "main",
		Channel: "ch-1",
		Address: "127.0.0.1:9000",
	}
	if sub.Cmd != uint8(2) || sub.ID != "sub-1" || sub.Route != "main" ||
		sub.Channel != "ch-1" || sub.Address != "127.0.0.1:9000" {
		t.Errorf("Subscriber fields not assigned correctly: %+v", sub)
	}
}

func TestMessageJSONTag(t *testing.T) {
	msg := Message{
		Cmd:    uint8(1),
		ID:     "id1",
		Route:  "route1",
		Status: StatusResolved,
		Body:   []byte("some payload"),
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	jsonStr := string(jsonBytes)
	encodedBody := base64.StdEncoding.EncodeToString(msg.Body)

	expectedFields := []string{
		`"cmd":1`,
		`"id":"id1"`,
		`"route":"route1"`,
		fmt.Sprintf(`"status":%d`, int(StatusResolved)),
		fmt.Sprintf(`"body":"%s"`, encodedBody),
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain: %s\nGot: %s", field, jsonStr)
		}
	}
}
