package commands

import (
	"encoding/json"
	"testing"
)

func TestStatusConstants(t *testing.T) {
	if StatusCreated != 0 {
		t.Errorf("StatusCreated expected 0, got %d", StatusCreated)
	}
	if StatusPending != 1 {
		t.Errorf("StatusPending expected 1, got %d", StatusPending)
	}
	if StatusResolved != 2 {
		t.Errorf("StatusResolved expected 2, got %d", StatusResolved)
	}
	if StatusInvalid != 3 {
		t.Errorf("StatusInvalid expected 3, got %d", StatusInvalid)
	}
}

func TestSendMessageStruct(t *testing.T) {
	msg := SendMessage{
		ID:     "123",
		Route:  "main",
		Status: StatusCreated,
		Body:   "Hello!",
	}

	if msg.ID != "123" || msg.Route != "main" || msg.Status != StatusCreated ||
		msg.Body != "Hello!" {
		t.Errorf("SendMessage fields not assigned correctly: %+v", msg)
	}
}

func TestAddSubscriberStruct(t *testing.T) {
	sub := AddSubscriber{
		ID:      "sub-1",
		Route:   "main",
		Channel: "ch-1",
		Address: "127.0.0.1:9000",
	}
	if sub.ID != "sub-1" || sub.Route != "main" || sub.Channel != "ch-1" ||
		sub.Address != "127.0.0.1:9000" {
		t.Errorf("AddSubscriber fields not assigned correctly: %+v", sub)
	}
}

func TestSendMessageJSONTag(t *testing.T) {
	msg := SendMessage{
		ID:     "id1",
		Route:  "route1",
		Status: StatusResolved,
		Body:   "some payload",
	}

	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	jsonStr := string(jsonBytes)
	expectedFields := []string{
		`"id":"id1"`,
		`"route":"route1"`,
		`"status":2`,
		`"body":"some payload"`,
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain: %s\nGot: %s", field, jsonStr)
		}
	}
}

// Helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
