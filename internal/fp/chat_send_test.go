package fp

import (
	"os"
	"strings"
	"testing"
)

func TestEncodeChatMessageBody(t *testing.T) {
	body, err := encodeChatMessageBody("users-4759067-16950672", 4919215462, "фыв", "tok123")
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	wantContains := []string{
		"csrf_token=tok123",
		"%22node%22%3A%22users-4759067-16950672%22",
		"%22last_message%22%3A4919215462",
		"%22content%22%3A%22",
		"%22action%22%3A%22chat_message%22",
	}
	for _, w := range wantContains {
		if !strings.Contains(body, w) {
			t.Errorf("body missing %q\ngot: %s", w, body)
		}
	}
}

func TestParseSendMessageResponse(t *testing.T) {
	body, err := os.ReadFile("../../scratch/send-message-response.txt")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	msgID, err := parseSendMessageResponse(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if msgID != 4920436371 {
		t.Errorf("message id: got %d, want 4920436371", msgID)
	}
}
