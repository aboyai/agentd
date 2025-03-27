// tools/search_tool.go
package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// WebSearch runs a basic DuckDuckGo search and extracts top links
func WebSearch(query string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	escaped := strings.ReplaceAll(query, " ", "+")
	url := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", escaped)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Sprintf("[error: %v]", err)
	}

	req.Header.Set("User-Agent", "AgentD/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf("[error: %v]", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Sprintf("[parse error: %v]", err)
	}

	var results []string
	doc.Find("a.result__a").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			text := s.Text()
			results = append(results, fmt.Sprintf("- %s (%s)", text, href))
			if len(results) >= 5 {
				return
			}
		}
	})

	if len(results) == 0 {
		return "[no results found]"
	}
	return strings.Join(results, "\n")
}
