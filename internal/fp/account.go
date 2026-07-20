package fp

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

func parseBalanceAmount(s string) (decimal.Decimal, error) {
	s = strings.ReplaceAll(s, "\u00A0", " ")
	s = strings.TrimSpace(s)

	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) || r == ',' || r == '.' || r == ' ' {
			b.WriteRune(r)
		}
	}
	s = b.String()

	lastComma := strings.LastIndex(s, ",")
	lastDot := strings.LastIndex(s, ".")
	switch {
	case lastComma > lastDot:
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	case lastDot > lastComma:
		s = strings.ReplaceAll(s, ",", "")
	}

	s = strings.ReplaceAll(s, " ", "")

	if s == "" {
		return decimal.Decimal{}, fmt.Errorf("empty balance after normalize")
	}
	balance, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("parse %q: %w", s, err)
	}
	return balance, nil
}

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
	balance, err := parseBalanceAmount(balanceText)
	if err != nil {
		return Account{}, fmt.Errorf("parse balance: %q: %w", balanceText, err)
	}

	return Account{
		UserID:  userID,
		Login:   login,
		Balance: balance,
	}, nil
}
