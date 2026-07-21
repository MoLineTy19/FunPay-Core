package fp

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseChatMessagesHTML(t *testing.T) {
	body, err := os.ReadFile("../../scratch/runner-read-1-response.txt")
	if err != nil {
		t.Skipf("sample not found: %v", err)
	}

	resp, err := decodeRunner(body)
	if err != nil {
		t.Fatalf("decodeRunner: %v", err)
	}

	objs, err := decodeRunnerObjects(resp.Objects)
	if err != nil {
		t.Fatalf("decodeRunnerObjects: %v", err)
	}

	var bookmark runnerObject
	for _, o := range objs {
		if o.Type == "chat_bookmarks" {
			bookmark = o
			break
		}
	}
	if bookmark.Type == "" {
		t.Fatal("chat_bookmarks object not found in fixture")
	}

	var d dataHTML
	if err := json.Unmarshal(bookmark.Data, &d); err != nil {
		t.Fatalf("unmarshal chat_bookmarks data: %v", err)
	}
	if d.HTML == "" {
		t.Fatal("data.html is empty in fixture")
	}

	out, err := ParseChatMessagesHTML(d.HTML)
	if err != nil {
		t.Fatalf("ParseChatMessagesHTML: %v", err)
	}

	if len(out) != 1 {
		t.Fatalf("expected 1 message, got %d", len(out))
	}
	msg := out[0]
	if msg.Text != "Как дела" {
		t.Errorf("Text: got %q, want %q", msg.Text, "Как дела")
	}
	if msg.Author != AuthorBuyer {
		t.Errorf("Author: got %q, want %q", msg.Author, AuthorBuyer)
	}
	if msg.ChatID != "274346432" {
		t.Errorf("ChatID: got %q, want %q", msg.ChatID, "274346432")
	}
	if msg.ID != "4906295429" {
		t.Errorf("ID: got %q, want %q", msg.ID, "4906295429")
	}

	if msg.CreatedAt.Hour() != 11 || msg.CreatedAt.Minute() != 12 {
		t.Errorf("CreatedAt hour:minute: got %02d:%02d, want 11:12",
			msg.CreatedAt.Hour(), msg.CreatedAt.Minute())
	}
}
