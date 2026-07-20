package fp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
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

	for _, f := range schema.Fields {
		value, ok := fields[f.ID]
		if !ok {
			continue
		} // поля нет во входных — пропускаем
		switch f.Type {
		case FieldText, FieldImages:
			if f.Type == FieldImages {
				continue
			} // images skip
			v.Set("fields["+f.ID+"]", value)
		case FieldMultilingual, FieldTextarea:
			v.Set("fields["+f.ID+"][ru]", value)
			v.Set("fields["+f.ID+"][en]", value)
		}
	}

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
