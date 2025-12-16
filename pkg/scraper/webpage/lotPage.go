package webpage

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractFormSelects парсит HTML-страницу и извлекает все элементы <select> из формы,
// расположенной в блоке .page-content.
func ExtractFormSelects(html []byte) (map[string]map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))

	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)

	doc.Find(".page-content").Find("form").ChildrenFiltered("div").Each(func(i int, block *goquery.Selection) {
		if block.Find("select").Length() > 0 {
			selectTag := block.Find("select")

			if selectTag.Length() == 0 {
				return // нет <select> в этом <div> — пропускаем
			}

			selectName, isExists := selectTag.Attr("name")
			if !isExists || selectName == "" {
				return // нет имени — пропускаем
			}

			// Создаём карту опций для этого <select>
			if _, ok := result[selectName]; !ok {
				result[selectName] = make(map[string]string)
			}

			optionsMap := result[selectName]

			block.Find("option").Each(func(j int, option *goquery.Selection) {
				value, hasValue := option.Attr("value")

				text := strings.TrimSpace(option.Text())

				if !hasValue || value == "" {
					return // пропускаем опции без <value> или с пустым
				}

				optionsMap[value] = text
			})
		}
	})

	return result, nil
}
