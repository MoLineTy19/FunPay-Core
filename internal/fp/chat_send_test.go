package fp

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestEncodeChatMessageBodyBrowserFormat(t *testing.T) {
	params := ChatMessageParams{
		UserID:       "16950672",
		Node:         "users-4759067-16950672",
		LastMessage:  4923043566,
		Text:         "фыв",
		CSRFToken:    "jsp4uiarn8ea5e35",
		OrdersTag:    "b72trv51",
		NodeTag:      "2vdj47fo",
		BookmarksTag: "ncs64qrg",
		Bookmarks:    [][]int64{{274346432, 4923043566}, {204298120, 4906267154}},
		CPUID:        "4759067",
		CPUTag:       "b9tqxiy3",
	}
	body, err := encodeChatMessageBody(params)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	vals, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("parse body as query: %v", err)
	}
	if got := vals.Get("csrf_token"); got != "jsp4uiarn8ea5e35" {
		t.Errorf("csrf_token: got %q, want jsp4uiarn8ea5e35", got)
	}

	var objs []map[string]json.RawMessage
	if err := json.Unmarshal([]byte(vals.Get("objects")), &objs); err != nil {
		t.Fatalf("parse objects: %v\nobjects=%s", err, vals.Get("objects"))
	}
	if len(objs) != 4 {
		t.Fatalf("objects count: got %d, want 4", len(objs))
	}

	want := []struct {
		typeVal string
		id      string
		tag     string
	}{
		{"orders_counters", "16950672", "b72trv51"},
		{"chat_node", "users-4759067-16950672", "2vdj47fo"},
		{"chat_bookmarks", "16950672", "ncs64qrg"},
		{"c-p-u", "4759067", "b9tqxiy3"},
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

	var bm [][]int64
	json.Unmarshal(objs[2]["data"], &bm)
	if len(bm) != 2 || bm[0][0] != 274346432 || bm[0][1] != 4923043566 {
		t.Errorf("chat_bookmarks data: got %+v, want [[274346432 4923043566] [...]]", bm)
	}

	type chatNodeData struct {
		Node        string `json:"node"`
		LastMessage int64  `json:"last_message"`
		Content     string `json:"content"`
	}
	var nodeData chatNodeData
	json.Unmarshal(objs[1]["data"], &nodeData)
	if nodeData.Node != "users-4759067-16950672" {
		t.Errorf("node data.node: got %q", nodeData.Node)
	}
	if nodeData.LastMessage != 4923043566 {
		t.Errorf("node data.last_message: got %d, want 4923043566", nodeData.LastMessage)
	}
	if nodeData.Content != "фыв" {
		t.Errorf("node data.content: got %q, want фыв", nodeData.Content)
	}

	for i, o := range objs {
		var tv string
		json.Unmarshal(o["type"], &tv)
		if tv == "chat_counter" {
			t.Errorf("obj[%d] chat_counter must NOT be present", i)
		}
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
}

func TestEncodeChatMessageBodyEmptyTagsAllowed(t *testing.T) {
	params := ChatMessageParams{
		UserID: "16950672", Node: "users-1-2", LastMessage: 0, Text: "hi", CSRFToken: "csrf",
	}
	body, err := encodeChatMessageBody(params)
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
