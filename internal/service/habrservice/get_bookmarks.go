package habrservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type HabrAuthor struct {
	ID       string `json:"id"`
	Alias    string `json:"alias"`
	FullName string `json:"fullname"`
}

type HabrHub struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
	Title string `json:"title"`
}

type HabrArticle struct {
	ID            string     `json:"id"`
	TimePublished string     `json:"timePublished"`
	Title         string     `json:"titleHtml"`
	Author        HabrAuthor `json:"author"`
	Hubs          []HabrHub  `json:"hubs"`
	LeadData      *HabrLead  `json:"leadData"`
	TextHtml      string     `json:"textHtml"`
	ReadingTime   int        `json:"readingTime"`
	Complexity    string     `json:"complexity"`
	TotalVotes    int        `json:"statistics"`
}

type HabrLead struct {
	TextHtml string `json:"textHtml"`
	Image    *struct {
		URL string `json:"url"`
	} `json:"image"`
}

type HabrBookmark struct {
	ID        int         `json:"id"`
	Alias     string      `json:"alias"`
	CreatedAt string      `json:"createdAt"`
	Article   HabrArticle `json:"article"`
}

type HabrBookmarksResponse struct {
	BookmarkIds  []int                   `json:"bookmarkIds"`
	BookmarkRefs map[string]HabrBookmark `json:"bookmarkRefs"`
	ArticleRefs  map[string]HabrArticle  `json:"articleRefs"`
}

type ParsedHabrArticle struct {
	ID          string
	Title       string
	Description string
	URL         string
	Hubs        []string
}

func (h *HabrService) GetBookmarks(ctx context.Context, page int) ([]ParsedHabrArticle, error) {
	if h.authToken == "" {
		return nil, fmt.Errorf("habr auth token is not set")
	}

	url := fmt.Sprintf("%s/bookmarks?page=%d", habrAPIURL, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("habr_session=%s", h.authToken))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get bookmarks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var bookmarksResp HabrBookmarksResponse
	if err := json.Unmarshal(body, &bookmarksResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return h.parseBookmarks(bookmarksResp), nil
}

func (h *HabrService) parseBookmarks(resp HabrBookmarksResponse) []ParsedHabrArticle {
	var articles []ParsedHabrArticle

	for _, bookmarkID := range resp.BookmarkIds {
		bookmark, ok := resp.BookmarkRefs[strconv.Itoa(bookmarkID)]
		if !ok {
			continue
		}

		article := bookmark.Article

		var hubs []string
		for _, hub := range article.Hubs {
			hubs = append(hubs, hub.Title)
		}

		description := ""
		if article.LeadData != nil {
			description = article.LeadData.TextHtml
		}

		articles = append(articles, ParsedHabrArticle{
			ID:          article.ID,
			Title:       article.Title,
			Description: stripHTML(description),
			URL:         fmt.Sprintf("%s/articles/%s/", habrBaseURL, article.ID),
			Hubs:        hubs,
		})
	}

	return articles
}

func stripHTML(html string) string {
	// Simple HTML stripping - for production use a proper library
	result := ""
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result += string(r)
		}
	}
	return result
}
