package fp

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestDecodeRunnerSmoke(t *testing.T) {
	body, err := os.ReadFile("../../scratch/runner-read-1-response.txt")
	if err != nil {
		t.Skipf("sample not found: %v", err)
	}

	resp, err := decodeRunner(body)
	if err != nil {
		t.Fatalf("decodeRunner: %v", err)
	}
	fmt.Printf("response=%v, %d objects\n", resp.Response, len(resp.Objects))

	objs, err := decodeRunnerObjects(resp.Objects)
	if err != nil {
		t.Fatalf("decodeRunnerObjects: %v", err)
	}
	for i, o := range objs {
		fmt.Printf("[%d] type=%s tag=%s\n", i, o.Type, o.Tag)
		if o.Type == "chat_counter" {
			d, err := decodeChatCounter(o)
			if err != nil {
				t.Fatalf("decodeChatCounter: %v", err)
			}
			fmt.Printf("    counter=%d message=%d\n", d.Counter, d.Message)
		}
	}
}

func TestChatBookmarksSmoke(t *testing.T) {
	data, err := os.ReadFile("../../scratch/chat.html")
	if err != nil {
		t.Skipf("sample not found: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	bookmarksTag, ok := doc.Find(".chat[data-bookmarks-tag]").Attr("data-bookmarks-tag")
	if !ok {
		t.Fatal("data-bookmarks-tag not found on /chat/")
	}
	fmt.Printf("bookmarksTag=%s\n", bookmarksTag)

	out := []chatBookmark{}
	doc.Find(".contact-item").Each(func(i int, s *goquery.Selection) {
		chatIDStr, ok1 := s.Attr("data-id")
		msgIDStr, ok2 := s.Attr("data-node-msg")
		if !ok1 || !ok2 {
			return
		}
		chatID, _ := strconv.ParseInt(chatIDStr, 10, 64)
		msgID, _ := strconv.ParseInt(msgIDStr, 10, 64)
		out = append(out, chatBookmark{ChatID: chatID, LastMessageID: msgID})
	})

	fmt.Printf("found %d bookmarks:\n", len(out))
	for i, b := range out {
		fmt.Printf("  [%d] chatID=%d lastMessageID=%d\n", i, b.ChatID, b.LastMessageID)
	}
}

func TestEncodeRunnerSmoke(t *testing.T) {
	// те же объекты, что в scratch/runner-payload.txt (декодировано)
	objs := []runnerRequestObject{
		{Type: "orders_counters", ID: "16950672", Tag: "8jeuxei9", Data: false},
		{Type: "chat_counter", ID: "16950672", Tag: "tvycophi", Data: false},
	}
	body, err := encodeRunnerRequest(objs, "n5val3ybryh2jt5v", false)
	if err != nil {
		t.Fatalf("encodeRunnerRequest: %v", err)
	}
	fmt.Printf("ENCODED BODY:\n%s\n", string(body))
}

