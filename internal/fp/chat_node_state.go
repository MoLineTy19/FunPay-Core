package fp

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func parseChatNodeState(body []byte) (lastMessage int64, nodeTag string, err error) {
	doc, perr := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if perr != nil {
		return 0, "", fmt.Errorf("parse html: %w", perr)
	}

	sel := doc.Find(".contact-item.active")
	if sel.Length() == 0 {
		return 0, "", fmt.Errorf("active contact-item not found")
	}

	msgStr, ok := sel.Attr("data-node-msg")
	if !ok {
		return 0, "", fmt.Errorf("data-node-msg not found on active contact-item")
	}
	lastMessage, err = strconv.ParseInt(msgStr, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("parse data-node-msg %q: %w", msgStr, err)
	}

	chatSel := doc.Find(".chat[data-tag]")
	chatSel.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		name, has := s.Attr("data-name")
		if !has || name == "" {
			return true
		}
		if t, ok2 := s.Attr("data-tag"); ok2 && t != "" {
			nodeTag = t
			return false
		}
		return true
	})

	return lastMessage, nodeTag, nil
}
