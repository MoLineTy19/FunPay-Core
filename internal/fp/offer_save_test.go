package fp

import (
	"errors"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestEncodeOfferForm(t *testing.T) {
	schema := OfferSchema{
		NodeID: "791",
		Fields: []OfferField{
			{ID: "level", Type: FieldText},
			{ID: "summary", Type: FieldMultilingual},
			{ID: "desc", Type: FieldTextarea},
			{ID: "images", Type: FieldImages},
		},
	}
	fields := map[string]string{
		"level":   "111",
		"summary": "Test Lot",
		"desc":    "desc text",
	}
	price, _ := decimal.NewFromString("111111")

	v := encodeOfferForm("csrf-token-123", "791", schema, fields, price, 5, true)

	// Base-ключи.
	checks := map[string]string{
		"csrf_token": "csrf-token-123",
		"offer_id":   "0",
		"node_id":    "791",
		"price":      "111111",
		"active":     "on",
		"amount":     "5",
	}
	for key, want := range checks {
		if got := v.Get(key); got != want {
			t.Errorf("form[%q]: got %q, want %q", key, got, want)
		}
	}

	// form_created_at — непустой numeric.
	fca := v.Get("form_created_at")
	if fca == "" || !strings.ContainsAny(fca, "0123456789") {
		t.Errorf("form_created_at: %q invalid", fca)
	}

	// FieldText: fields[level] = "111".
	if v.Get("fields[level]") != "111" {
		t.Errorf("fields[level]: got %q, want 111", v.Get("fields[level]"))
	}

	// FieldMultilingual: fields[summary][ru] = fields[summary][en] = "Test Lot".
	if v.Get("fields[summary][ru]") != "Test Lot" {
		t.Errorf("fields[summary][ru]: got %q", v.Get("fields[summary][ru]"))
	}
	if v.Get("fields[summary][en]") != "Test Lot" {
		t.Errorf("fields[summary][en]: got %q", v.Get("fields[summary][en]"))
	}

	// FieldTextarea: оба языка.
	if v.Get("fields[desc][ru]") != "desc text" {
		t.Errorf("fields[desc][ru]: got %q", v.Get("fields[desc][ru]"))
	}
	if v.Get("fields[desc][en]") != "desc text" {
		t.Errorf("fields[desc][en]: got %q", v.Get("fields[desc][en]"))
	}

	// FieldImages: skip — поля fields[images] быть не должно.
	if _, ok := v["fields[images]"]; ok {
		t.Errorf("fields[images]: should be skipped (FieldImages type)")
	}
}

func TestEncodeOfferFormExtraFieldsIgnored(t *testing.T) {
	// Входные fields содержат ключ, которого нет в схеме — игнорируется.
	schema := OfferSchema{
		NodeID: "791",
		Fields: []OfferField{{ID: "summary", Type: FieldMultilingual}},
	}
	fields := map[string]string{
		"summary": "Test",
		"unknown": "should be ignored",
	}
	price := decimal.NewFromInt(100)

	v := encodeOfferForm("csrf", "791", schema, fields, price, 0, true)

	if _, ok := v["fields[unknown]"]; ok {
		t.Errorf("fields[unknown]: should be ignored (not in schema)")
	}
	if _, ok := v["fields[unknown][ru]"]; ok {
		t.Errorf("fields[unknown][ru]: should be ignored (not in schema)")
	}
}

func TestEncodeOfferFormAmountZero(t *testing.T) {
	// amount=0 → шлём пустым (бесконечный stock).
	schema := OfferSchema{NodeID: "791", Fields: []OfferField{{ID: "summary", Type: FieldMultilingual}}}
	fields := map[string]string{"summary": "Test"}
	price := decimal.NewFromInt(100)

	v := encodeOfferForm("csrf", "791", schema, fields, price, 0, true)
	if v.Get("amount") != "" {
		t.Errorf("amount=0: got %q, want empty", v.Get("amount"))
	}
}

func TestParseSaveResponseOK(t *testing.T) {
	body := []byte(`{
		"done": true,
		"error": false,
		"errors": [],
		"url": "https://funpay.com/lots/791/trade"
	}`)
	resp, err := parseSaveResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Done {
		t.Errorf("Done: got false, want true")
	}
	if len(resp.Errors) != 0 {
		t.Errorf("Errors: got %v, want empty", resp.Errors)
	}
	if resp.URL != "https://funpay.com/lots/791/trade" {
		t.Errorf("URL: got %q", resp.URL)
	}
}

func TestParseSaveResponseValidation(t *testing.T) {
	// errors — список пар [field, msg] (реверс FPC: for k,v in errors).
	body := []byte(`{"done":false,"error":true,"errors":[["price","required"],["summary","too short"]],"url":""}`)
	resp, err := parseSaveResponse(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Done {
		t.Errorf("Done: got true, want false")
	}
	if len(resp.Errors) != 2 {
		t.Fatalf("Errors: got %d, want 2", len(resp.Errors))
	}
	// Первая пара: ["price", "required"].
	if len(resp.Errors[0]) != 2 || resp.Errors[0][0] != "price" || resp.Errors[0][1] != "required" {
		t.Errorf("Errors[0]: got %v, want [price required]", resp.Errors[0])
	}
}

func TestParseSaveResponseAuthLost(t *testing.T) {
	body := []byte(`{"msg":"cookie expired","error":1}`)
	_, err := parseSaveResponse(body)
	if !errors.Is(err, ErrAuthLost) {
		t.Fatalf("err: got %v, want ErrAuthLost", err)
	}
}
