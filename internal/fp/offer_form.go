package fp

import (
	"bytes"
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
	NodeID string
	Fields []OfferField
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

	return OfferSchema{NodeID: nodeID, Fields: fields}, nil
}
