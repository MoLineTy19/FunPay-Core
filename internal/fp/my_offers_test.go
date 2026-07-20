package fp

import (
	"os"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseMyOffers(t *testing.T) {
	body, err := os.ReadFile("../../scratch/lots-791-trade.html")
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	offers, err := parseMyOffers(body, "791")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(offers) != 2 {
		t.Fatalf("offers count: got %d, want 2", len(offers))
	}
	first := offers[0]
	if first.OfferID != "73311257" {
		t.Errorf("first.OfferID: got %q, want 73311257", first.OfferID)
	}
	if first.NodeID != "791" {
		t.Errorf("first.NodeID: got %q, want 791", first.NodeID)
	}
	if first.Server != "Android" {
		t.Errorf("first.Server: got %q, want Android", first.Server)
	}
	wantSummary := "12312, 111111 уровень, 12111 глава"
	if first.Summary != wantSummary {
		t.Errorf("first.Summary: got %q, want %q", first.Summary, wantSummary)
	}
	wantPrice, _ := decimal.NewFromString("111111")
	if !first.Price.Equal(wantPrice) {
		t.Errorf("first.Price: got %s, want %s", first.Price, wantPrice)
	}
	if offers[1].OfferID != "73146557" {
		t.Errorf("second.OfferID: got %q, want 73146557", offers[1].OfferID)
	}
}

func TestParseMyOffersEmpty(t *testing.T) {
	body := []byte(`<html><body><div class="content-lots-trade">Нет предложений</div></body></html>`)
	offers, err := parseMyOffers(body, "791")
	if err != nil {
		t.Fatalf("parse empty: %v", err)
	}
	if len(offers) != 0 {
		t.Errorf("offers count: got %d, want 0", len(offers))
	}
}
