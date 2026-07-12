package vkservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type VKFaveItem struct {
	AddedDate int    `json:"added_date"`
	Type      string `json:"type"`
	Post      *struct {
		ID          int    `json:"id"`
		FromID      int    `json:"from_id"`
		OwnerID     int    `json:"owner_id"`
		Text        string `json:"text"`
		Attachments []struct {
			Type    string `json:"type"`
			Article *struct {
				ID      int    `json:"id"`
				OwnerID int    `json:"owner_id"`
				Title   string `json:"title"`
				URL     string `json:"url"`
			} `json:"article"`
			Link *struct {
				URL         string `json:"url"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"link"`
		} `json:"attachments"`
	} `json:"post"`
	Link *struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"link"`
	Article *struct {
		ID      int    `json:"id"`
		OwnerID int    `json:"owner_id"`
		Title   string `json:"title"`
		URL     string `json:"url"`
	} `json:"article"`
}

type VKFavesResponse struct {
	Response struct {
		Count int          `json:"count"`
		Items []VKFaveItem `json:"items"`
	} `json:"response"`
	Error *struct {
		ErrorCode int    `json:"error_code"`
		ErrorMsg  string `json:"error_msg"`
	} `json:"error"`
}

type ParsedFave struct {
	SourceID string
	Title    string
	Body     string
	URL      string
	Type     string
}

func (v *VkService) GetFaves(ctx context.Context, offset, count int) ([]ParsedFave, int, error) {
	if v.accessToken == "" {
		return nil, 0, fmt.Errorf("vk access token is not set")
	}

	uri := vkAPIBaseURL + "/fave.get"
	form := url.Values{
		"access_token": {v.accessToken},
		"v":            {vkAPIVersion},
		"count":        {strconv.Itoa(count)},
		"offset":       {strconv.Itoa(offset)},
		"extended":     {"1"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.URL.RawQuery = form.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("get faves: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read body: %w", err)
	}

	var faves VKFavesResponse
	if err := json.Unmarshal(body, &faves); err != nil {
		return nil, 0, fmt.Errorf("decode response: %w", err)
	}

	if faves.Error != nil {
		return nil, 0, fmt.Errorf("vk api error %d: %s", faves.Error.ErrorCode, faves.Error.ErrorMsg)
	}

	parsed := v.parseFaves(faves.Response.Items)
	return parsed, faves.Response.Count, nil
}

func (v *VkService) parseFaves(items []VKFaveItem) []ParsedFave {
	var result []ParsedFave

	for _, item := range items {
		switch item.Type {
		case "article":
			if item.Article != nil {
				result = append(result, ParsedFave{
					SourceID: fmt.Sprintf("article_%d_%d", item.Article.OwnerID, item.Article.ID),
					Title:    item.Article.Title,
					URL:      item.Article.URL,
					Type:     "article",
				})
			}
		case "link":
			if item.Link != nil {
				result = append(result, ParsedFave{
					SourceID: fmt.Sprintf("link_%s", hashURL(item.Link.URL)),
					Title:    item.Link.Title,
					Body:     item.Link.Description,
					URL:      item.Link.URL,
					Type:     "link",
				})
			}
		case "post":
			if item.Post != nil {
				for _, att := range item.Post.Attachments {
					switch att.Type {
					case "article":
						if att.Article != nil {
							result = append(result, ParsedFave{
								SourceID: fmt.Sprintf("article_%d_%d", att.Article.OwnerID, att.Article.ID),
								Title:    att.Article.Title,
								URL:      att.Article.URL,
								Type:     "article",
							})
						}
					case "link":
						if att.Link != nil {
							result = append(result, ParsedFave{
								SourceID: fmt.Sprintf("link_%s", hashURL(att.Link.URL)),
								Title:    att.Link.Title,
								Body:     att.Link.Description,
								URL:      att.Link.URL,
								Type:     "link",
							})
						}
					}
				}
				if item.Post.Text != "" && len(item.Post.Attachments) == 0 {
					result = append(result, ParsedFave{
						SourceID: fmt.Sprintf("post_%d_%d", item.Post.OwnerID, item.Post.ID),
						Title:    truncate(item.Post.Text, 100),
						Body:     item.Post.Text,
						URL:      fmt.Sprintf("https://vk.com/wall%d_%d", item.Post.OwnerID, item.Post.ID),
						Type:     "post",
					})
				}
			}
		}
	}

	return result
}

func hashURL(u string) string {
	h := uint32(0)
	for _, c := range u {
		h = h*31 + uint32(c)
	}
	return strconv.FormatUint(uint64(h), 16)
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
