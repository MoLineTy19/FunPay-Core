package fp

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestEncodeChatMessageBodyBrowserFormat(t *testing.T) {
	body, err := encodeChatMessageBody(
		"16950672",
		"users-4759067-16950672",
		4919215462,
		"фыв",
		"lkv4iivvnymfni18",
		"hywmot8v",
		"wo7l5k7g",
		"alk772vd",
	)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	vals, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("parse body as query: %v", err)
	}
	if got := vals.Get("csrf_token"); got != "lkv4iivvnymfni18" {
		t.Errorf("csrf_token: got %q, want lkv4iivvnymfni18", got)
	}

	var objs []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(vals.Get("objects")), &objs); err != nil {
		t.Fatalf("parse objects: %v\nobjects=%s", err, vals.Get("objects"))
	}
	if len(objs) != 3 {
		t.Fatalf("objects count: got %d, want 3", len(objs))
	}

	want := []struct {
		typeVal string
		id      string
		tag     string
	}{
		{"orders_counters", "16950672", "hywmot8v"},
		{"chat_counter", "16950672", "wo7l5k7g"},
		{"chat_node", "users-4759067-16950672", "alk772vd"},
	}
	for i, w := range want {
		var tv, id, tag string
		json.Unmarshal(objs[i]["type"], &tv)
		json.Unmarshal(objs[i]["id"], &id)
		json.Unmarshal(objs[i]["tag"], &tag)
		if tv != w.typeVal {
			t.Errorf("obj[%d] type: got %q, want %q", i, tv, w.typeVal)
		}
		if id != w.id {
			t.Errorf("obj[%d] id: got %q, want %q", i, id, w.id)
		}
		if tag != w.tag {
			t.Errorf("obj[%d] tag: got %q, want %q", i, tag, w.tag)
		}
	}

	type chatNodeData struct {
		Node        string `json:"node"`
		LastMessage int64  `json:"last_message"`
		Content     string `json:"content"`
	}
	var nodeData chatNodeData
	json.Unmarshal(objs[2]["data"], &nodeData)
	if nodeData.Node != "users-4759067-16950672" {
		t.Errorf("node data.node: got %q, want users-4759067-16950672", nodeData.Node)
	}
	if nodeData.LastMessage != 4919215462 {
		t.Errorf("node data.last_message: got %d, want 4919215462", nodeData.LastMessage)
	}
	if nodeData.Content != "фыв" {
		t.Errorf("node data.content: got %q, want фыв", nodeData.Content)
	}

	var req struct {
		Action string       `json:"action"`
		Data   chatNodeData `json:"data"`
	}
	if err := json.Unmarshal([]byte(vals.Get("request")), &req); err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if req.Action != "chat_message" {
		t.Errorf("request.action: got %q, want chat_message", req.Action)
	}
	if req.Data.Node != "users-4759067-16950672" {
		t.Errorf("request.data.node: got %q", req.Data.Node)
	}
}

func TestEncodeChatMessageBodyEmptyTagsAllowed(t *testing.T) {
	body, err := encodeChatMessageBody(
		"16950672", "users-1-2", 0, "hi", "csrf", "", "", "",
	)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(body, "%22tag%22%3A%22%22") {
		t.Errorf("empty tags should produce empty tag values; got: %s", body)
	}
	if !strings.Contains(body, "%22last_message%22%3A0") {
		t.Errorf("last_message 0 expected; got: %s", body)
	}
}

func TestParseSendMessageResponse(t *testing.T) {
	body, err := os.ReadFile("../../scratch/send-message-response.txt")
	if err != nil {
		t.Skipf("sample missing: %v", err)
	}
	sent, err := parseSendMessageResponse(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if sent.MessageID != 4920436371 {
		t.Errorf("message id: got %d, want 4920436371", sent.MessageID)
	}
	if sent.NodeTag != "whespj30" {
		t.Errorf("node tag: got %q, want whespj30", sent.NodeTag)
	}
}
