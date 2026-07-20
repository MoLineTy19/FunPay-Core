package fp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type saveOfferResponse struct {
	Done   bool       `json:"done"`
	Error  bool       `json:"error"`
	Errors [][]string `json:"errors"`
	URL    string     `json:"url"`
}

func parseSaveResponse(body []byte) (saveOfferResponse, error) {
	var re runnerError
	if json.Unmarshal(body, &re) == nil && re.Error != 0 {
		return saveOfferResponse{}, fmt.Errorf("%w: %s", ErrAuthLost, re.Msg)
	}
	var resp saveOfferResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return saveOfferResponse{}, fmt.Errorf("decode saveOfferResponse: %w", err)
	}
	return resp, nil
}

func encodeOfferForm(csrfToken, nodeID string, schema OfferSchema, fields map[string]string, price decimal.Decimal, amount int, active bool) url.Values {
	v := url.Values{}
	v.Set("csrf_token", csrfToken)
	v.Set("form_created_at", strconv.FormatInt(time.Now().Unix(), 10))
	v.Set("offer_id", "0")
	v.Set("node_id", nodeID)
	v.Set("location", "")
	v.Set("deleted", "")
	if schema.ServerID != "" {
		v.Set("server_id", schema.ServerID)
	}

	for _, f := range schema.Fields {
		switch f.Type {
		case FieldText:
			value, ok := fields[f.ID]
			if !ok {
				continue
			}
			v.Set("fields["+f.ID+"]", value)
		case FieldMultilingual, FieldTextarea:
			value, ok := fields[f.ID]
			if !ok {
				continue
			}
			v.Set("fields["+f.ID+"][ru]", value)
			v.Set("fields["+f.ID+"][en]", value)
		case FieldImages:
			v.Set("fields["+f.ID+"]", "")
		}
	}

	v.Set("secrets", "")
	v.Set("price", price.String())
	if amount > 0 {
		v.Set("amount", strconv.Itoa(amount))
	} else {
		v.Set("amount", "")
	}
	if active {
		v.Set("active", "on")
	}
	return v
}

type OfferCreated struct {
	NodeID  string
	OfferID string
	URL     string
}

// CreateOffer создаёт лот на FP. Три шага:
//  1. GetOfferForm(nodeID) — узнаём схему полей категории.
//  2. encodeOfferForm + POST /lots/offerSave → parseSaveResponse (ловит auth-lost и валидацию).
//  3. GetMyOffers(nodeID) + match по fields["summary"] → OfferID нового лота.
//
// FP не возвращает ID нового лота в ответе offerSave (только {done, url}), поэтому
// нужен шаг 3 — поиск по описанию в списке моих лотов.
func (c *Client) CreateOffer(ctx context.Context, csrfToken, nodeID string, fields map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error) {
	schema, err := c.GetOfferForm(ctx, nodeID)
	if err != nil {
		return OfferCreated{}, fmt.Errorf("get form schema: %w", err)
	}

	form := encodeOfferForm(csrfToken, nodeID, schema, fields, price, amount, active)
	body, err := c.do(ctx, "POST", "https://funpay.com/lots/offerSave",
		strings.NewReader(form.Encode()),
		"application/x-www-form-urlencoded; charset=UTF-8")
	if err != nil {
		return OfferCreated{}, fmt.Errorf("offerSave request: %w", err)
	}

	resp, err := parseSaveResponse(body)
	if err != nil {
		return OfferCreated{}, err // уже ErrAuthLost wrapped
	}
	if !resp.Done || len(resp.Errors) > 0 {
		return OfferCreated{}, fmt.Errorf("offerSave validation failed: %v", resp.Errors)
	}

	summary := fields["summary"]
	offers, err := c.GetMyOffers(ctx, nodeID)
	if err != nil {
		return OfferCreated{}, fmt.Errorf("find created offer: %w", err)
	}
	for _, o := range offers {
		if o.Summary == summary {
			return OfferCreated{NodeID: nodeID, OfferID: o.OfferID, URL: resp.URL}, nil
		}
	}
	return OfferCreated{}, fmt.Errorf("created offer not found in /lots/%s/trade (summary=%q)", nodeID, summary)
}
