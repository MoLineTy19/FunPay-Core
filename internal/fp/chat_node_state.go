package fp

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type ChatNodeState struct {
	LastMessage int64
	NodeTag     string
	CPUID       string
	CPUTag      string
}

func parseChatNodeState(body []byte) (ChatNodeState, error) {
	doc, perr := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if perr != nil {
		return ChatNodeState{}, fmt.Errorf("parse html: %w", perr)
	}

	sel := doc.Find(".contact-item.active")
	if sel.Length() == 0 {
		return ChatNodeState{}, fmt.Errorf("active contact-item not found")
	}

	var st ChatNodeState

	msgStr, ok := sel.Attr("data-node-msg")
	if !ok {
		return ChatNodeState{}, fmt.Errorf("data-node-msg not found on active contact-item")
	}
	st.LastMessage, perr = strconv.ParseInt(msgStr, 10, 64)
	if perr != nil {
		return ChatNodeState{}, fmt.Errorf("parse data-node-msg %q: %w", msgStr, perr)
	}

	doc.Find(".chat[data-tag]").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		name, has := s.Attr("data-name")
		if !has || name == "" {
			return true
		}
		if t, ok2 := s.Attr("data-tag"); ok2 && t != "" {
			st.NodeTag = t
			return false
		}
		return true
	})

	doc.Find(`.param-item[data-type="c-p-u"]`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if id, ok2 := s.Attr("data-id"); ok2 && id != "" {
			st.CPUID = id
		}
		if t, ok2 := s.Attr("data-tag"); ok2 && t != "" {
			st.CPUTag = t
		}
		return false
	})

	return st, nil
}
