package fp

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

func parseSalesOrders(body []byte) ([]OrderShortcut, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var out []OrderShortcut
	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		id := extractOrderIDFromHref(href)
		if id == "" {
			return
		}

		class, _ := s.Attr("class")
		status := StatusNew
		switch {
		case strings.Contains(class, "warning"):
			status = StatusCancelled
		case strings.Contains(class, "info"):
			status = StatusNew
		default:
			status = StatusCompleted
		}

		summary := strings.TrimSpace(s.Find(".order-desc").Find("div").First().Text())
		buyerName := strings.TrimSpace(s.Find(".media-user-name").Text())

		priceText := strings.TrimSpace(s.Find(".tc-seller-sum").Text())
		priceText = strings.SplitN(priceText, " ", 2)[0]
		price, perr := decimal.NewFromString(priceText)
		if perr != nil {
			price = decimal.Zero
		}

		out = append(out, OrderShortcut{
			ID:        id,
			Status:    status,
			BuyerName: buyerName,
			Summary:   summary,
			Price:     price,
		})
	})
	return out, nil
}

func extractOrderIDFromHref(href string) string {
	const prefix = "https://funpay.com/orders/"
	const suffix = "/"
	if !strings.HasPrefix(href, prefix) || !strings.HasSuffix(href, suffix) {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(href, prefix), suffix)
}

func (c *Client) GetSales(ctx context.Context) ([]OrderShortcut, error) {
	data, err := c.do(ctx, "GET", "https://funpay.com/orders/trade", nil, "")
	if err != nil {
		return nil, fmt.Errorf("get sales: %w", err)
	}
	return parseSalesOrders(data)
}
