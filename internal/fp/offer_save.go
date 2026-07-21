package fp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

func encodeOfferForm(nodeID, serverID string, schema OfferSchema, fields map[string]string, price decimal.Decimal, amount int, active bool) url.Values {
	v := url.Values{}
	v.Set("csrf_token", schema.CSRFToken)
	v.Set("form_created_at", schema.FormCreatedAt)
	v.Set("offer_id", "0")
	v.Set("node_id", nodeID)
	v.Set("location", "")
	v.Set("deleted", "")
	if serverID != "" {
		v.Set("server_id", serverID)
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
//  1. GetOfferForm(nodeID) — схема полей + csrf_token + form_created_at из формы.
//  2. encodeOfferForm + POST /lots/offerSave → parseSaveResponse (ловит auth-lost и валидацию).
//  3. GetMyOffers(nodeID) + match по fields["summary"] → OfferID нового лота.
//
// FP не возвращает ID нового лота в ответе offerSave (только {done, url}), поэтому
func (c *Client) CreateOffer(ctx context.Context, nodeID, serverID string, fields map[string]string, price decimal.Decimal, amount int, active bool) (OfferCreated, error) {
	schema, err := c.GetOfferForm(ctx, nodeID)
	if err != nil {
		return OfferCreated{}, fmt.Errorf("get form schema: %w", err)
	}

	form := encodeOfferForm(nodeID, serverID, schema, fields, price, amount, active)
	referer := "https://funpay.com/lots/offerEdit?node=" + nodeID
	body, err := c.doWithReferer(ctx, "POST", "https://funpay.com/lots/offerSave",
		strings.NewReader(form.Encode()),
		"application/x-www-form-urlencoded; charset=UTF-8", referer)
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
		// FP формирует .tc-desc-text = summary + level + stage, поэтому точное
		// равенство не сработает — матчим по вхождению summary.
		if strings.Contains(o.Summary, summary) {
			return OfferCreated{NodeID: nodeID, OfferID: o.OfferID, URL: resp.URL}, nil
		}
	}
	return OfferCreated{}, fmt.Errorf("created offer not found in /lots/%s/trade (summary=%q)", nodeID, summary)
}

// encodeOfferEditForm строит payload для edit-offer.
// Стартует с копии values.FieldValues (текущие значения), накладывает patch fields,
// и для price/amount/active: nil → оставляем текущее, не-nil → перетираем.
func encodeOfferEditForm(values LotValues, fields map[string]string, price *decimal.Decimal, amount *int, active *bool) url.Values {
	v := url.Values{}
	for k, val := range values.FieldValues {
		v.Set(k, val)
	}

	// Наложение патча fields. Тип определяется по ключу в FieldValues:
	// если ключ есть как fields[id][ru] → multilingual (оба языка), иначе text.
	for id, val := range fields {
		if id == "images" {
			continue // fields[images] не переопределяем из patch
		}
		ruKey := "fields[" + id + "][ru]"
		if _, isMulti := values.FieldValues[ruKey]; isMulti {
			v.Set(ruKey, val)
			v.Set("fields["+id+"][en]", val)
		} else {
			v.Set("fields["+id+"]", val)
		}
	}

	// Служебные поля.
	v.Set("offer_id", values.OfferID)
	v.Set("node_id", values.NodeID)
	v.Set("csrf_token", values.CSRFToken)
	v.Set("form_created_at", values.FormCreatedAt)
	if values.ServerID != "" {
		v.Set("server_id", values.ServerID)
	}
	v.Set("location", "")
	v.Set("deleted", "")

	// price / amount / active — слияние current+override.
	if price != nil {
		v.Set("price", price.String())
	}
	if amount != nil {
		if *amount > 0 {
			v.Set("amount", strconv.Itoa(*amount))
		} else {
			v.Set("amount", "")
		}
	}
	if active != nil {
		if *active {
			v.Set("active", "on")
		} else {
			v.Del("active")
		}
	}
	return v
}

// encodeDeleteOfferForm строит payload для delete-offer.
// Стартует с копии values.FieldValues (FP требует полный payload), ставит deleted=1.
func encodeDeleteOfferForm(values LotValues) url.Values {
	v := url.Values{}
	for k, val := range values.FieldValues {
		v.Set(k, val)
	}
	v.Set("offer_id", values.OfferID)
	v.Set("node_id", values.NodeID)
	v.Set("csrf_token", values.CSRFToken)
	v.Set("form_created_at", values.FormCreatedAt)
	if values.ServerID != "" {
		v.Set("server_id", values.ServerID)
	}
	v.Set("location", "")
	v.Set("deleted", "1")
	return v
}
