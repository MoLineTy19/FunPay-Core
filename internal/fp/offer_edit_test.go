package fp

import (
	"errors"
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseOfferEditForm(t *testing.T) {
	body, err := os.ReadFile("../../scratch/offer-edit-form-791.html")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	v, err := parseOfferEditForm(body, "791", "73311257")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v.OfferID != "73311257" {
		t.Errorf("OfferID: got %q, want 73311257", v.OfferID)
	}
	if v.NodeID != "791" {
		t.Errorf("NodeID: got %q, want 791", v.NodeID)
	}
	if v.ServerID != "5188" {
		t.Errorf("ServerID: got %q, want 5188", v.ServerID)
	}
	if v.CSRFToken != "0xt3wcb31x6tq9yd" {
		t.Errorf("CSRFToken: got %q, want 0xt3wcb31x6tq9yd", v.CSRFToken)
	}
	if v.FormCreatedAt != "1784569239" {
		t.Errorf("FormCreatedAt: got %q, want 1784569239", v.FormCreatedAt)
	}
	wantFields := map[string]string{
		"fields[level]":           "111111",
		"fields[stage]":           "12111",
		"fields[summary][ru]":     "12312",
		"fields[summary][en]":     "12312",
		"fields[desc][ru]":        "1231",
		"fields[desc][en]":        "ada",
		"fields[payment_msg][ru]": "",
		"fields[payment_msg][en]": "",
		"fields[images]":          "",
		"price":                   "111111",
		"amount":                  "1",
		"secrets":                 "",
		"location":                "",
		"deleted":                 "",
		"auto_delivery":           "",
	}
	for k, want := range wantFields {
		if got := v.FieldValues[k]; got != want {
			t.Errorf("FieldValues[%q]: got %q, want %q", k, got, want)
		}
	}
	if v.Amount != "1" {
		t.Errorf("Amount: got %q, want 1", v.Amount)
	}
	if v.Active {
		t.Errorf("Active: got true, want false (checkbox без [checked] в образце)")
	}
}

func TestParseOfferEditFormNotFound(t *testing.T) {
	body := []byte(`<html><body><p class="lead">Lot not found</p></body></html>`)
	_, err := parseOfferEditForm(body, "791", "999")
	if !errors.Is(err, ErrOfferNotFound) {
		t.Fatalf("err: got %v, want ErrOfferNotFound", err)
	}
}

func TestEncodeOfferEditForm(t *testing.T) {
	values := LotValues{
		NodeID:        "791",
		OfferID:       "73311257",
		ServerID:      "5188",
		CSRFToken:     "csrf-edit-xyz",
		FormCreatedAt: "1700000000",
		FieldValues: map[string]string{
			"fields[level]":           "111111",
			"fields[stage]":           "12111",
			"fields[summary][ru]":     "OLD_RU",
			"fields[summary][en]":     "OLD_EN",
			"fields[desc][ru]":        "old desc ru",
			"fields[desc][en]":        "old desc en",
			"fields[payment_msg][ru]": "",
			"fields[payment_msg][en]": "",
			"fields[images]":          "",
			"price":                   "111111",
			"amount":                  "1",
			"secrets":                 "",
			"location":                "",
			"deleted":                 "",
			"auto_delivery":           "",
		},
	}
	patch := map[string]map[string]string{"summary": {"ru": "NEW", "en": "NEW"}}
	price200, _ := decimal.NewFromString("200")
	amount3 := 3
	activeTrue := true

	v := encodeOfferEditForm(values, patch, &price200, &amount3, &activeTrue)

	checks := map[string]string{
		"offer_id":            "73311257",
		"node_id":             "791",
		"csrf_token":          "csrf-edit-xyz",
		"form_created_at":     "1700000000",
		"server_id":           "5188",
		"location":            "",
		"deleted":             "",
		"price":               "200",
		"amount":              "3",
		"active":              "on",
		"fields[summary][ru]": "NEW",
		"fields[summary][en]": "NEW",
		"fields[level]":       "111111",
		"fields[stage]":       "12111",
		"fields[desc][ru]":    "old desc ru",
		"fields[images]":      "",
		"secrets":             "",
		"auto_delivery":       "",
	}
	for k, want := range checks {
		if got := v.Get(k); got != want {
			t.Errorf("form[%q]: got %q, want %q", k, got, want)
		}
	}
}

func TestEncodeDeleteOfferForm(t *testing.T) {
	values := LotValues{
		NodeID:        "791",
		OfferID:       "73311257",
		ServerID:      "5188",
		CSRFToken:     "csrf-edit-xyz",
		FormCreatedAt: "1700000000",
		FieldValues: map[string]string{
			"fields[level]":       "111111",
			"fields[summary][ru]": "OLD",
			"fields[images]":      "",
			"price":               "111111",
			"amount":              "1",
			"secrets":             "",
			"location":            "",
			"deleted":             "",
		},
	}
	v := encodeDeleteOfferForm(values)

	if v.Get("deleted") != "1" {
		t.Errorf("deleted: got %q, want 1", v.Get("deleted"))
	}
	if v.Get("offer_id") != "73311257" {
		t.Errorf("offer_id: got %q", v.Get("offer_id"))
	}
	if v.Get("csrf_token") != "csrf-edit-xyz" {
		t.Errorf("csrf_token: got %q", v.Get("csrf_token"))
	}
	if v.Get("form_created_at") != "1700000000" {
		t.Errorf("form_created_at: got %q", v.Get("form_created_at"))
	}
	if v.Get("node_id") != "791" {
		t.Errorf("node_id: got %q", v.Get("node_id"))
	}
	if v.Get("server_id") != "5188" {
		t.Errorf("server_id: got %q", v.Get("server_id"))
	}
	if v.Get("location") != "" {
		t.Errorf("location: got %q, want empty", v.Get("location"))
	}
	if v.Get("fields[level]") != "111111" {
		t.Errorf("fields[level] should be preserved from values: got %q", v.Get("fields[level]"))
	}
	if v.Get("price") != "111111" {
		t.Errorf("price should be preserved: got %q", v.Get("price"))
	}
}

func TestEncodeOfferEditFormNilPointers(t *testing.T) {
	values := LotValues{
		NodeID:        "791",
		OfferID:       "73311257",
		ServerID:      "5188",
		CSRFToken:     "csrf",
		FormCreatedAt: "1",
		FieldValues: map[string]string{
			"fields[summary][ru]": "OLD",
			"price":               "999",
			"amount":              "5",
		},
	}
	v := encodeOfferEditForm(values, map[string]map[string]string{"summary": {"ru": "X"}}, nil, nil, nil)

	if v.Get("fields[summary][ru]") != "X" {
		t.Errorf("patched summary[ru]: got %q, want X", v.Get("fields[summary][ru]"))
	}
	if v.Get("price") != "999" {
		t.Errorf("price should be preserved when nil: got %q, want 999", v.Get("price"))
	}
	if v.Get("amount") != "5" {
		t.Errorf("amount should be preserved when nil: got %q, want 5", v.Get("amount"))
	}
	if _, ok := v["active"]; ok {
		t.Errorf("active should be absent when nil and not in values")
	}
}
