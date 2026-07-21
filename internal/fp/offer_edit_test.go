package fp

import (
	"errors"
	"os"
	"testing"
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
