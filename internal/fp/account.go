package fp

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

var userIDRe = regexp.MustCompile(`/users/(\d+)`)

func (c *Client) GetAccount(ctx context.Context) (Account, error) {
	data, err := c.do(ctx, "GET", "https://funpay.com/account/balance", nil, "")
	if err != nil {
		return Account{}, fmt.Errorf("get account: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(data))
	if err != nil {
		return Account{}, fmt.Errorf("parse html: %w", err)
	}

	login := doc.Find(".user-link-name").First().Text()

	href, _ := doc.Find(".user-link-dropdown").First().Attr("href")
	m := userIDRe.FindStringSubmatch(href)
	var userID int64
	if m != nil {
		userID, err = strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return Account{}, fmt.Errorf("parse userID: %w", err)
		}
	}

	balanceText := doc.Find(".balances-list .balances-value").First().Text()
	fields := strings.Fields(balanceText)
	var balance decimal.Decimal
	if len(fields) > 0 {
		balance, err = decimal.NewFromString(fields[0])
		if err != nil {
			return Account{}, fmt.Errorf("parse balance: %w", err)
		}
	} else {
		return Account{}, fmt.Errorf("balance element not found")
	}

	return Account{
		UserID:  userID,
		Login:   login,
		Balance: balance,
	}, nil
}
