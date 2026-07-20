package fp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"

	"github.com/PuerkitoBio/goquery"
)

type FieldType int

const (
	FieldText         FieldType = 1 // fields[id] — простое текстовое (level, stage)
	FieldMultilingual FieldType = 2 // fields[id][ru/en] — краткое multilingual (summary)
	FieldTextarea     FieldType = 3 // fields[id][ru/en] — длинное multilingual (desc, payment_msg)
	FieldImages       FieldType = 6 // fields[id] — hidden, картинки (skip при кодировке)
)

type OfferField struct {
	ID         string    `json:"id"`
	Type       FieldType `json:"type"`
	Conditions []any     `json:"conditions"`
}

type OfferSchema struct {
	NodeID        string
	ServerID      string
	CSRFToken     string
	FormCreatedAt string
	Fields        []OfferField
}

func parseOfferFormSchema(body []byte, nodeID string) (OfferSchema, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return OfferSchema{}, fmt.Errorf("parse html: %w", err)
	}
	selection := doc.Find(".lot-fields")
	if selection.Length() == 0 {
		return OfferSchema{}, fmt.Errorf("lot-fields element not found")
	}
	raw, exists := selection.Attr("data-fields")
	if !exists || raw == "" {
		return OfferSchema{}, fmt.Errorf("data-fields attribute not found")
	}

	// raw содержит HTML-entity-закодированный JSON (" вместо ").
	decoded := html.UnescapeString(raw)

	var fields []OfferField
	if err := json.Unmarshal([]byte(decoded), &fields); err != nil {
		return OfferSchema{}, fmt.Errorf("decode data-fields JSON: %w", err)
	}

	serverID, _ := doc.Find(`select[name="server_id"] option[selected]`).Attr("value")

	csrfToken, _ := doc.Find(`input[name="csrf_token"]`).Attr("value")
	formCreatedAt, _ := doc.Find(`input[name="form_created_at"]`).Attr("value")

	return OfferSchema{
		NodeID:        nodeID,
		ServerID:      serverID,
		CSRFToken:     csrfToken,
		FormCreatedAt: formCreatedAt,
		Fields:        fields,
	}, nil
}

func (c *Client) GetOfferForm(ctx context.Context, nodeID string) (OfferSchema, error) {
	url := "https://funpay.com/lots/offerEdit?node=" + nodeID
	data, err := c.do(ctx, "GET", url, nil, "")
	if err != nil {
		return OfferSchema{}, fmt.Errorf("get offer form: %w", err)
	}
	return parseOfferFormSchema(data, nodeID)
}
