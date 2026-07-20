package fp

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

type MyOffer struct {
	OfferID string
	NodeID  string
	Summary string          // .tc-desc-text целиком
	Server  string          // .tc-server-inside
	Amount  string          // .tc-amount (пусто = бесконечный stock)
	Price   decimal.Decimal // из data-s атрибута .tc-price
}

func parseMyOffers(body []byte, nodeID string) ([]MyOffer, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var out []MyOffer
	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		offerID, exists := s.Attr("data-offer")
		if !exists || offerID == "" {
			return
		}
		summary := strings.TrimSpace(s.Find(".tc-desc-text").Text())
		server := strings.TrimSpace(s.Find(".tc-server-inside").Text())
		amount := strings.TrimSpace(s.Find(".tc-amount").Text())
		priceStr, _ := s.Find(".tc-price").Attr("data-s")
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			price = decimal.Zero
		}
		out = append(out, MyOffer{
			OfferID: offerID,
			NodeID:  nodeID,
			Summary: summary,
			Server:  server,
			Amount:  amount,
			Price:   price,
		})
	})
	return out, nil
}

func (c *Client) GetMyOffers(ctx context.Context, nodeID string) ([]MyOffer, error) {
	url := "https://funpay.com/lots/" + nodeID + "/trade"
	data, err := c.do(ctx, "GET", url, nil, "")
	if err != nil {
		return nil, fmt.Errorf("get my offers: %w", err)
	}
	return parseMyOffers(data, nodeID)
}
