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
	st, err := parseChatNodeState(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if st.LastMessage != 4922342027 {
		t.Errorf("last message: got %d, want 4922342027", st.LastMessage)
	}
	if st.NodeTag != "0g3dit1v" {
		t.Errorf("node tag: got %q, want 0g3dit1v", st.NodeTag)
	}
	if st.CPUID != "4759067" {
		t.Errorf("cpu id: got %q, want 4759067", st.CPUID)
	}
	if st.CPUTag != "b9tqxiy3" {
		t.Errorf("cpu tag: got %q, want b9tqxiy3", st.CPUTag)
	}
}

func TestParseChatNodeStateNoActive(t *testing.T) {
	body := []byte(`<html><body><a class="contact-item" data-id="1" data-node-msg="100"></a></body></html>`)
	_, err := parseChatNodeState(body)
	if err == nil {
		t.Fatal("want error for missing active contact, got nil")
	}
}
