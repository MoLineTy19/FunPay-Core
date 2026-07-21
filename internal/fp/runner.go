package fp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

var ErrAuthLost = errors.New("auth lost: golden_seal expired or missing")

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

type runnerError struct {
	Msg   string `json:"msg"`
	Error int    `json:"error"`
}

type runnerChatCounterData struct {
	Counter int   `json:"counter"`
	Message int64 `json:"message"`
}

type runnerOrdersCountersData struct {
	Buyer  int `json:"buyer"`
	Seller int `json:"seller"`
}

func diffOrderSnapshots(prev map[string]OrderShortcut, current []OrderShortcut) []OrderEvent {
	events := make([]OrderEvent, 0, len(current))
	for _, sc := range current {
		old, exists := prev[sc.ID]
		if !exists {
			events = append(events, OrderEvent{
				Order:    sc,
				Kind:     OrderEventNew,
				ToStatus: sc.Status,
			})
			continue
		}
		if old.Status == sc.Status {
			continue
		}
		kind := OrderEventKind("")
		switch sc.Status {
		case StatusCompleted:
			kind = OrderEventCompleted
		case StatusCancelled:
			kind = OrderEventCancelled
		}
		if kind == "" {
			continue
		}
		events = append(events, OrderEvent{
			Order:      sc,
			Kind:       kind,
			FromStatus: old.Status,
			ToStatus:   sc.Status,
		})
	}
	return events
}

type runnerRequestObject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Tag  string `json:"tag"`
	Data any    `json:"data"`
}

type Runner struct {
	client      *Client
	userID      string
	csrfToken   string
	objectTypes []string
	tags        map[string]string
	bookmarks   []chatBookmark
	orders      map[string]OrderShortcut
}

type RunnerEvents struct {
	Messages []ChatMessage
	Orders   []OrderEvent
}

type dataHTML struct {
	HTML string `json:"html"`
}

type chatBookmark struct {
	ChatID         int64
	LastMessageID  int64
	LastUserReadID int64
}

func decodeRunner(body []byte) (runnerResponse, error) {
	resp := runnerResponse{}

	if len(body) == 0 {
		return runnerResponse{}, nil
	}

	var re runnerError
	if err := json.Unmarshal(body, &re); err == nil && re.Error != 0 {
		return runnerResponse{}, fmt.Errorf("%w: %s", ErrAuthLost, re.Msg)
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
		bookmarks:   nil,
		orders:      make(map[string]OrderShortcut),
	}
}

func (r *Runner) Poll(ctx context.Context) (RunnerEvents, error) {
	events := RunnerEvents{}

	objs := make([]runnerRequestObject, 0, len(r.objectTypes))
	for _, t := range r.objectTypes {
		obj := runnerRequestObject{
			Type: t,
			ID:   r.userID,
			Tag:  r.tags[t],
		}

		if t == "chat_bookmarks" {
			data := make([][]int64, 0, len(r.bookmarks))
			for _, b := range r.bookmarks {
				if b.LastMessageID == b.LastUserReadID {
					data = append(data, []int64{b.ChatID, b.LastMessageID})
				} else {
					data = append(data, []int64{b.ChatID, b.LastMessageID, b.LastUserReadID})
				}
			}
			obj.Data = data
		} else {
			obj.Data = false
		}

		objs = append(objs, obj)
	}

	req, err := encodeRunnerRequest(objs, r.csrfToken, false)
	if err != nil {
		return RunnerEvents{}, fmt.Errorf("encode runner request: %w", err)
	}

	res, err := r.client.do(ctx, "POST", "https://funpay.com/runner/", bytes.NewReader(req), "application/x-www-form-urlencoded; charset=UTF-8")
	if err != nil {
		return RunnerEvents{}, fmt.Errorf("execute runner: %w", err)
	}

	runnerResp, err := decodeRunner(res)
	if err != nil {
		return RunnerEvents{}, fmt.Errorf("decode runner response: %w", err)
	}

	runner, err := decodeRunnerObjects(runnerResp.Objects)
	if err != nil {
		return RunnerEvents{}, fmt.Errorf("decode runner object response: %w", err)
	}

	for _, obj := range runner {
		r.tags[obj.Type] = obj.Tag
		switch obj.Type {
		case "chat_bookmarks":
			var d dataHTML
			if err := json.Unmarshal(obj.Data, &d); err != nil {
				return RunnerEvents{}, fmt.Errorf("decode chat_bookmarks html: %w", err)
			}
			msgs, err := ParseChatMessagesHTML(d.HTML)
			if err != nil {
				return RunnerEvents{}, fmt.Errorf("parse chat_bookmarks html: %w", err)
			}
			events.Messages = append(events.Messages, msgs...)
		case "orders_counters":
			evs, err := r.diffOrders(ctx, obj)
			if err != nil {
				return RunnerEvents{}, fmt.Errorf("diff orders: %w", err)
			}
			events.Orders = append(events.Orders, evs...)
		}
	}

	return events, nil
}

func (r *Runner) diffOrders(ctx context.Context, obj runnerObject) ([]OrderEvent, error) {
	var data runnerOrdersCountersData
	if err := json.Unmarshal(obj.Data, &data); err != nil {
		return nil, fmt.Errorf("decode orders_counters data: %w", err)
	}
	if data.Buyer == 0 && data.Seller == 0 {
		return nil, nil
	}
	current, err := r.client.GetSales(ctx)
	if err != nil {
		return nil, fmt.Errorf("get sales: %w", err)
	}
	events := diffOrderSnapshots(r.orders, current)
	r.orders = make(map[string]OrderShortcut, len(current))
	for _, o := range current {
		r.orders[o.ID] = o
	}
	return events, nil
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

	bookmarks, bookmarksTag, err := getChatBookmarks(ctx, r.client)
	if err != nil {
		return fmt.Errorf("runner init: %w", err)
	}

	r.bookmarks = bookmarks
	r.tags["chat_bookmarks"] = bookmarksTag

	initialOrders, err := loadInitialOrders(ctx, r.client)
	if err != nil {
		return fmt.Errorf("runner init: %w", err)
	}
	r.orders = initialOrders
	return nil
}

func loadInitialOrders(ctx context.Context, client *Client) (map[string]OrderShortcut, error) {
	current, err := client.GetSales(ctx)
	if err != nil {
		return nil, fmt.Errorf("load initial orders: %w", err)
	}
	out := make(map[string]OrderShortcut, len(current))
	for _, o := range current {
		out[o.ID] = o
	}
	return out, nil
}

func getChatBookmarks(ctx context.Context, client *Client) ([]chatBookmark, string, error) {
	data, err := client.do(ctx, "GET", "https://funpay.com/chat/", nil, "")
	if err != nil {
		return nil, "", fmt.Errorf("get chat bookmarks: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("parse chat bookmarks: %w", err)
	}

	bookmarksTag, ok := doc.Find(".chat[data-bookmarks-tag]").Attr("data-bookmarks-tag")
	if !ok {
		return nil, "", fmt.Errorf("data-bookmarks-tag not found on /chat/")
	}

	out := []chatBookmark{}
	doc.Find(".contact-item").Each(func(i int, s *goquery.Selection) {
		chatIDStr, ok1 := s.Attr("data-id")
		msgIDStr, ok2 := s.Attr("data-node-msg")
		userMsgStr, ok3 := s.Attr("data-user-msg")
		if !ok1 || !ok2 || !ok3 {
			return
		}

		chatID, err1 := strconv.ParseInt(chatIDStr, 10, 64)
		msgID, err2 := strconv.ParseInt(msgIDStr, 10, 64)
		userMsg, err3 := strconv.ParseInt(userMsgStr, 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			return
		}

		out = append(out, chatBookmark{
			ChatID:         chatID,
			LastMessageID:  msgID,
			LastUserReadID: userMsg,
		})
	})

	return out, bookmarksTag, nil
}
