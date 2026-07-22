package fp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type MessageSent struct {
	MessageID int64
	NodeTag   string
	Raw       json.RawMessage
}

type ChatMessageParams struct {
	UserID       string
	Node         string
	LastMessage  int64
	Text         string
	CSRFToken    string
	OrdersTag    string
	NodeTag      string
	BookmarksTag string
	Bookmarks    [][]int64
	CPUID        string
	CPUTag       string
}

func encodeChatMessageBody(p ChatMessageParams) (string, error) {
	type chatNodeData struct {
		Node        string `json:"node"`
		LastMessage int64  `json:"last_message"`
		Content     string `json:"content"`
	}
	type runnerObj struct {
		Type string      `json:"type"`
		ID   string      `json:"id"`
		Tag  string      `json:"tag"`
		Data interface{} `json:"data"`
	}
	type requestData struct {
		Action string       `json:"action"`
		Data   chatNodeData `json:"data"`
	}

	data := chatNodeData{Node: p.Node, LastMessage: p.LastMessage, Content: p.Text}
	objs := []runnerObj{
		{Type: "orders_counters", ID: p.UserID, Tag: p.OrdersTag, Data: false},
		{Type: "chat_node", ID: p.Node, Tag: p.NodeTag, Data: data},
		{Type: "chat_bookmarks", ID: p.UserID, Tag: p.BookmarksTag, Data: p.Bookmarks},
		{Type: "c-p-u", ID: p.CPUID, Tag: p.CPUTag, Data: false},
	}
	req := requestData{Action: "chat_message", Data: data}

	objsJSON, err := json.Marshal(objs)
	if err != nil {
		return "", fmt.Errorf("encode objects: %w", err)
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	v := url.Values{}
	v.Set("objects", string(objsJSON))
	v.Set("request", string(reqJSON))
	v.Set("csrf_token", p.CSRFToken)
	return v.Encode(), nil
}

func parseSendMessageResponse(body []byte) (MessageSent, error) {
	var raw struct {
		Objects []struct {
			Type string `json:"type"`
			Tag  string `json:"tag"`
			Data struct {
				Messages []struct {
					ID json.RawMessage `json:"id"`
				} `json:"messages"`
			} `json:"data"`
		} `json:"objects"`
		Response struct {
			Error json.RawMessage `json:"error"`
		} `json:"response"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return MessageSent{}, fmt.Errorf("parse send message response: %w", err)
	}
	if string(raw.Response.Error) != "null" && len(raw.Response.Error) > 0 {
		return MessageSent{}, fmt.Errorf("send message error: %s", string(raw.Response.Error))
	}
	for _, obj := range raw.Objects {
		if obj.Type != "chat_node" {
			continue
		}
		if len(obj.Data.Messages) == 0 {
			continue
		}
		var id int64
		if err := json.Unmarshal(obj.Data.Messages[0].ID, &id); err == nil {
			return MessageSent{MessageID: id, NodeTag: obj.Tag}, nil
		}
		var idStr string
		if err := json.Unmarshal(obj.Data.Messages[0].ID, &idStr); err == nil {
			parsed, _ := strconv.ParseInt(idStr, 10, 64)
			return MessageSent{MessageID: parsed, NodeTag: obj.Tag}, nil
		}
	}
	return MessageSent{}, fmt.Errorf("no chat_node message in response")
}
