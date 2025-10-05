package parse

import (
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
		price := s.Find(".tc-price").Text()
		userLink, exists := s.Find(".avatar-photo").Attr("data-href")
		if !exists {
			userLink = "-"
		}

		fmt.Println(strings.TrimSpace(title), strings.TrimSpace(description), strings.TrimSpace(price), userLink)
	})
}
