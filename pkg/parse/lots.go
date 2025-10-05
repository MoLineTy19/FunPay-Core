package parse

import (
	"FunPay-Core/pkg/types"
	"FunPay-Core/pkg/utils"
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetLotsData Извлечение данных из HTML-страницы
func GetLotsData(html []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))

	if err != nil {
		panic(err)
	}

	doc.Find(".tc-item").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".tc-server-inside").Text()
		description := s.Find(".tc-desc-text").Text()
		price := utils.ParsePrice(s.Find(".tc-price").Text())

		userLink, exists := s.Find(".avatar-photo").Attr("data-href")
		if !exists {
			userLink = "-"
		}
		userName := s.Find(".media-user-name").Text()
		lot := types.Lot{
			Title:       strings.TrimSpace(title),
			Description: strings.TrimSpace(description),
			Price:       price,
			UserLink:    userLink,
			UserName:    strings.TrimSpace(userName),
		}
		fmt.Println(lot)
	})
}
