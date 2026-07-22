package fp

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

func parseTradeTags(body []byte) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
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

	return map[string]string{
		"orders_counters": ordersTag,
		"chat_counter":    chatTag,
	}, nil
}
