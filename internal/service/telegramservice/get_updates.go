package telegramservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type TGUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type TGChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type TGMessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url"`
}

type TGMessage struct {
	MessageID int64             `json:"message_id"`
	From      *TGUser           `json:"from"`
	Chat      TGChat            `json:"chat"`
	Date      int64             `json:"date"`
	Text      string            `json:"text"`
	Entities  []TGMessageEntity `json:"entities"`
}

type TGUpdate struct {
	UpdateID int64      `json:"update_id"`
	Message  *TGMessage `json:"message"`
}

type TGUpdatesResponse struct {
	Ok     bool       `json:"ok"`
	Result []TGUpdate `json:"result"`
}

type ParsedMessage struct {
	MessageID int64
	ChatID    int64
	Text      string
	URLs      []string
	Date      int64
}

func (t *TelegramService) GetUpdates(ctx context.Context, offset int64, limit int) ([]TGUpdate, error) {
	if t.botToken == "" {
		return nil, fmt.Errorf("telegram bot token is not set")
	}

	uri := fmt.Sprintf("%s%s/getUpdates", tgAPIBaseURL, t.botToken)
	params := url.Values{
		"offset":  {strconv.FormatInt(offset, 10)},
		"limit":   {strconv.Itoa(limit)},
		"timeout": {"30"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var updates TGUpdatesResponse
	if err := json.Unmarshal(body, &updates); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !updates.Ok {
		return nil, fmt.Errorf("telegram api error")
	}

	return updates.Result, nil
}

func (t *TelegramService) ParseMessages(updates []TGUpdate) []ParsedMessage {
	var messages []ParsedMessage

	urlRegex := regexp.MustCompile(`https?://[^\s]+`)

	for _, update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		var urls []string

		// Extract URLs from entities
		for _, entity := range msg.Entities {
			if entity.Type == "url" {
				urlText := extractSubstring(msg.Text, entity.Offset, entity.Length)
				urls = append(urls, urlText)
			} else if entity.Type == "text_link" && entity.URL != "" {
				urls = append(urls, entity.URL)
			}
		}

		// Also find URLs in text using regex
		found := urlRegex.FindAllString(msg.Text, -1)
		for _, u := range found {
			if !contains(urls, u) {
				urls = append(urls, u)
			}
		}

		if len(urls) > 0 {
			messages = append(messages, ParsedMessage{
				MessageID: msg.MessageID,
				ChatID:    msg.Chat.ID,
				Text:      msg.Text,
				URLs:      urls,
				Date:      msg.Date,
			})
		}
	}

	return messages
}

func extractSubstring(s string, offset, length int) string {
	runes := []rune(s)
	if offset >= len(runes) {
		return ""
	}
	end := offset + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[offset:end])
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (t *TelegramService) ExtractTitle(text string, url string) string {
	// Try to use first line as title
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && len(lines[0]) > 0 {
		title := strings.TrimSpace(lines[0])
		// Remove URL from title if present
		title = strings.Replace(title, url, "", -1)
		title = strings.TrimSpace(title)
		if len(title) > 0 && len(title) <= 200 {
			return title
		}
	}

	// Fallback to URL
	return url
}
