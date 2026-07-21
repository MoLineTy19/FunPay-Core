package fp

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

func parseOrderDetail(body []byte) (Order, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return Order{}, fmt.Errorf("parse html: %w", err)
	}

	canonical, ok := doc.Find(`link[rel="canonical"]`).Attr("href")
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	id := extractOrderIDFromHref(canonical)
	if id == "" {
		return Order{}, ErrOrderNotFound
	}

	var buyerID int64
	var buyerName string
	doc.Find(".media-user-name a").Each(func(_ int, s *goquery.Selection) {
		if buyerID != 0 {
			return
		}
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		if bid := extractUserIDFromHref(href); bid != 0 {
			buyerID = bid
			buyerName = strings.TrimSpace(s.Text())
		}
	})

	chatID := ""
	doc.Find(`.chat-control[href*="node="]`).Each(func(_ int, s *goquery.Selection) {
		if chatID != "" {
			return
		}
		href, _ := s.Attr("href")
		chatID = extractQueryParam(href, "node")
	})

	nodeID := ""
	doc.Find(`.param-item a[href*="/lots/"]`).Each(func(_ int, s *goquery.Selection) {
		if nodeID != "" {
			return
		}
		href, _ := s.Attr("href")
		nodeID = extractNodeIDFromLotsHref(href)
	})

	amount, currency := extractAmountAndCurrency(doc)

	return Order{
		ID:        id,
		NodeID:    nodeID,
		BuyerID:   buyerID,
		BuyerName: buyerName,
		Amount:    amount,
		Currency:  currency,
		Status:    StatusNew,
		ChatID:    chatID,
	}, nil
}

func extractAmountAndCurrency(doc *goquery.Document) (decimal.Decimal, string) {
	var amount decimal.Decimal
	var currency string
	doc.Find(".param-item").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if !strings.Contains(s.Text(), "Сумма") {
			return true
		}
		amountText := strings.TrimSpace(s.Find(".h1").Text())
		if a, err := decimal.NewFromString(amountText); err == nil {
			amount = a
		}
		currency = strings.TrimSpace(s.Find("strong").Text())
		return false
	})
	return amount, currency
}

func extractUserIDFromHref(href string) int64 {
	re := regexp.MustCompile(`/users/(\d+)/`)
	m := re.FindStringSubmatch(href)
	if len(m) < 2 {
		return 0
	}
	id, _ := strconv.ParseInt(m[1], 10, 64)
	return id
}

func extractQueryParam(rawURL, key string) string {
	idx := strings.Index(rawURL, key+"=")
	if idx < 0 {
		return ""
	}
	rest := rawURL[idx+len(key)+1:]
	end := strings.IndexAny(rest, "&#")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func extractNodeIDFromLotsHref(href string) string {
	re := regexp.MustCompile(`/lots/(\d+)/`)
	m := re.FindStringSubmatch(href)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (Order, error) {
	url := "https://funpay.com/orders/" + orderID + "/"
	data, err := c.do(ctx, "GET", url, nil, "")
	if err != nil {
		return Order{}, fmt.Errorf("get order: %w", err)
	}
	return parseOrderDetail(data)
}
