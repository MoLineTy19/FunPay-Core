package fp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type runnerObject struct {
	Type string          `json:"type"`
	ID   json.RawMessage `json:"id"`
	Tag  string          `json:"tag"`
	Data json.RawMessage `json:"data"`
}

type runnerResponse struct {
	Objects  []json.RawMessage `json:"objects"`
	Response bool              `json:"response"`
}

type runnerChatCounterData struct {
	Counter int   `json:"counter"`
	Message int64 `json:"message"`
}

type runnerRequestObject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Tag  string `json:"tag"`
	Data bool   `json:"data"`
}

type Runner struct {
	client      *Client
	userID      string
	csrfToken   string
	objectTypes []string
	tags        map[string]string
}

func decodeRunner(body []byte) (runnerResponse, error) {
	resp := runnerResponse{}

	if len(body) == 0 {
		return runnerResponse{}, nil
	}

	err := json.Unmarshal(body, &resp)
	if err != nil {
		return runnerResponse{}, fmt.Errorf("decode runner: %w", err)
	}
	return resp, nil
}

func decodeRunnerObjects(raw []json.RawMessage) ([]runnerObject, error) {
	out := make([]runnerObject, 0, len(raw))

	for _, r := range raw {
		obj := runnerObject{}
		err := json.Unmarshal(r, &obj)
		if err != nil {
			log.Printf("decode object: %v", err)
			continue
		}
		out = append(out, obj)
	}
	return out, nil
}

func decodeChatCounter(obj runnerObject) (runnerChatCounterData, error) {
	if obj.Type != "chat_counter" {
		return runnerChatCounterData{}, fmt.Errorf("unexpected type %s", obj.Type)
	}

	data := runnerChatCounterData{}
	err := json.Unmarshal(obj.Data, &data)
	if err != nil {
		return runnerChatCounterData{}, fmt.Errorf("decode chat_counter data: %w", err)
	}

	return data, nil
}

func encodeRunnerRequest(objects []runnerRequestObject, csrfToken string, request bool) ([]byte, error) {
	raw, err := json.Marshal(objects)
	if err != nil {
		return nil, fmt.Errorf("encode runner objects: %w", err)
	}

	v := url.Values{}
	v.Set("objects", string(raw))
	v.Set("request", strconv.FormatBool(request))
	v.Set("csrf_token", csrfToken)

	return []byte(v.Encode()), nil
}

func NewRunner(client *Client, userID, csrfToken string, objectTypes []string) *Runner {
	tags := make(map[string]string)
	return &Runner{
		client:      client,
		userID:      userID,
		csrfToken:   csrfToken,
		objectTypes: objectTypes,
		tags:        tags,
	}
}

func (r *Runner) Poll(ctx context.Context) (runnerResponse, error) {
	objs := make([]runnerRequestObject, 0, len(r.objectTypes))
	for _, t := range r.objectTypes {
		objs = append(objs, runnerRequestObject{
			Type: t,
			ID:   r.userID,
			Tag:  r.tags[t],
			Data: false,
		})
	}

	req, err := encodeRunnerRequest(objs, r.csrfToken, false)
	if err != nil {
		return runnerResponse{}, fmt.Errorf("encode runner request: %w", err)
	}

	res, err := r.client.do(ctx, "POST", "https://funpay.com/runner/", bytes.NewReader(req), "application/x-www-form-urlencoded")
	if err != nil {
		return runnerResponse{}, fmt.Errorf("execute runner: %w", err)
	}

	runnerResp, err := decodeRunner(res)
	if err != nil {
		return runnerResponse{}, fmt.Errorf("decode runner response: %w", err)
	}

	runner, err := decodeRunnerObjects(runnerResp.Objects)
	if err != nil {
		return runnerResponse{}, fmt.Errorf("decode runner object response: %w", err)
	}

	for _, obj := range runner {
		r.tags[obj.Type] = obj.Tag
	}

	return runnerResponse{
		Objects:  runnerResp.Objects,
		Response: runnerResp.Response,
	}, nil
}

func getInitialTags(ctx context.Context, client *Client) (map[string]string, error) {
	data, err := client.do(ctx, "GET", "https://funpay.com/orders/trade", nil, "")
	if err != nil {
		return nil, fmt.Errorf("get initial tags: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse initial tags: %w", err)
	}

	sel := doc.Find("#live-counters")
	if sel.Length() == 0 {
		return nil, fmt.Errorf("live-counters element not found")
	}

	ordersTag, ok1 := sel.Attr("data-orders")
	chatTag, ok2 := sel.Attr("data-chat")

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("data-orders or data-chat attribute not found")
	}

	tags := map[string]string{
		"orders_counters": ordersTag,
		"chat_counter":    chatTag,
	}

	return tags, nil
}

func (r *Runner) Init(ctx context.Context) error {
	tags, err := getInitialTags(ctx, r.client)
	if err != nil {
		return fmt.Errorf("runner init: %w", err)
	}

	r.tags = tags
	return nil
}
