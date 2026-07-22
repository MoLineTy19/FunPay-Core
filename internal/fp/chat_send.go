package fp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type MessageSent struct {
	MessageID int64
	NodeTag   string
	Raw       json.RawMessage
}

func encodeChatMessageBody(userID, node string, lastMessage int64, text, csrfToken, ordersTag, chatTag, nodeTag string) (string, error) {
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

	data := chatNodeData{Node: node, LastMessage: lastMessage, Content: text}
	objs := []runnerObj{
		{Type: "orders_counters", ID: userID, Tag: ordersTag, Data: false},
		{Type: "chat_counter", ID: userID, Tag: chatTag, Data: false},
		{Type: "chat_node", ID: node, Tag: nodeTag, Data: data},
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
	v.Set("csrf_token", csrfToken)
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

func (c *Client) SendMessage(ctx context.Context, userID, node string, lastMessage int64, text, ordersTag, chatTag, nodeTag string) (MessageSent, error) {
	body, err := encodeChatMessageBody(userID, node, lastMessage, text, c.csrfToken, ordersTag, chatTag, nodeTag)
	if err != nil {
		return MessageSent{}, fmt.Errorf("encode chat message: %w", err)
	}
	resp, err := c.do(ctx, "POST", "https://funpay.com/runner/", strings.NewReader(body), "application/x-www-form-urlencoded; charset=UTF-8")
	if err != nil {
		return MessageSent{}, fmt.Errorf("send message: %w", err)
	}
	sent, err := parseSendMessageResponse(resp)
	if err != nil {
		return MessageSent{}, err
	}
	sent.Raw = resp
	return sent, nil
}
