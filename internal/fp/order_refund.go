package fp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type Refunded struct {
	OrderID string
	Raw     json.RawMessage
}

func encodeRefundBody(orderID, csrfToken string) string {
	v := url.Values{}
	v.Set("csrf_token", csrfToken)
	v.Set("id", orderID)
	return v.Encode()
}

func (c *Client) RefundOrder(ctx context.Context, orderID string) (Refunded, error) {
	body := encodeRefundBody(orderID, c.csrfToken)
	referer := "https://funpay.com/orders/" + orderID + "/"
	resp, err := c.doWithReferer(ctx, "POST", "https://funpay.com/orders/refund",
		strings.NewReader(body), "application/x-www-form-urlencoded; charset=UTF-8", referer)
	if err != nil {
		return Refunded{}, fmt.Errorf("refund order: %w", err)
	}
	return Refunded{OrderID: orderID, Raw: resp}, nil
}
