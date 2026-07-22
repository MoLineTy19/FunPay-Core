package fp

import (
	"os"
	"testing"
)

func TestParseChatNodeState(t *testing.T) {
	body, err := os.ReadFile("../../scratch/chat-node-active.html")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	lastMsg, nodeTag, err := parseChatNodeState(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if lastMsg != 4922342027 {
		t.Errorf("last message: got %d, want 4922342027", lastMsg)
	}
	if nodeTag != "0g3dit1v" {
		t.Errorf("node tag: got %q, want 0g3dit1v", nodeTag)
	}
}

func TestParseChatNodeStateNoActive(t *testing.T) {
	body := []byte(`<html><body><a class="contact-item" data-id="1" data-node-msg="100"></a></body></html>`)
	_, _, err := parseChatNodeState(body)
	if err == nil {
		t.Fatal("want error for missing active contact, got nil")
	}
}
