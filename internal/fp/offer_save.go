package fp

import (
	"net/url"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

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
