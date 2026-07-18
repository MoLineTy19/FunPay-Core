package fp

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ParseChatMessagesHTML(html string) ([]ChatMessage, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return []ChatMessage{}, fmt.Errorf("parse html: %w", err)
	}

	out := make([]ChatMessage, 0)
	doc.Find(".contact-item").Each(func(i int, s *goquery.Selection) {

		createdAt := s.Find(".contact-item-time").Text()
		msgTime, err := time.Parse("15:04", createdAt) // → 0001-01-01 11:12
		if err != nil {
			return
		}
		now := time.Now()
		createdAtObj := time.Date(now.Year(), now.Month(), now.Day(), msgTime.Hour(), msgTime.Minute(), 0, 0, now.Location())

		msgID, ok1 := s.Attr("data-node-msg")
		if !ok1 {
			return
		}
		chatID, ok2 := s.Attr("data-id")
		if !ok2 {
			return
		}

		out = append(out, ChatMessage{
			ID:        msgID,
			ChatID:    chatID,
			Text:      s.Find(".contact-item-message").Text(),
			Author:    AuthorBuyer,
			CreatedAt: createdAtObj,
		})
	})

	return out, nil
}
