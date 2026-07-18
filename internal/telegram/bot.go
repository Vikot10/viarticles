package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Vikot10/viarticles/internal/storage"
)

const apiBaseURL = "https://api.telegram.org"

var urlPattern = regexp.MustCompile(`https?://[^\s<>]+`)

type Bot struct {
	token         string
	allowedChatID int64
	store         *storage.Storage
	client        *http.Client
}

func New(token string, allowedChatID int64, store *storage.Storage) *Bot {
	return &Bot{
		token:         strings.TrimSpace(token),
		allowedChatID: allowedChatID,
		store:         store,
		client:        &http.Client{Timeout: 35 * time.Second},
	}
}

func (b *Bot) Run(ctx context.Context) {
	if b.token == "" {
		slog.Info("telegram bot disabled: token is empty")
		return
	}
	if b.allowedChatID == 0 {
		slog.Error("telegram bot disabled: VIARTICLES_TELEGRAM_ALLOWED_CHAT_ID is empty")
		return
	}

	offset := b.loadOffset(ctx)
	slog.Info("telegram bot started")
	for ctx.Err() == nil {
		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("telegram getUpdates failed", "error", err)
				wait(ctx, 3*time.Second)
			}
			continue
		}
		for _, update := range updates {
			if update.Message != nil {
				b.handleMessage(ctx, *update.Message)
			}
			offset = update.ID + 1
			if err := b.store.SetState(ctx, "telegram_offset", strconv.FormatInt(offset, 10)); err != nil {
				slog.Error("save telegram offset", "error", err)
			}
		}
	}
}

func (b *Bot) loadOffset(ctx context.Context) int64 {
	value, err := b.store.State(ctx, "telegram_offset")
	if err != nil {
		slog.Error("load telegram offset", "error", err)
		return 0
	}
	offset, _ := strconv.ParseInt(value, 10, 64)
	return offset
}

func (b *Bot) getUpdates(ctx context.Context, offset int64) ([]update, error) {
	var updates []update
	err := b.call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         25,
		"allowed_updates": []string{"message"},
	}, &updates)
	return updates, err
}

func (b *Bot) handleMessage(ctx context.Context, message message) {
	if message.Chat.ID != b.allowedChatID {
		return
	}
	text := message.Text
	if text == "" {
		text = message.Caption
	}
	if strings.HasPrefix(strings.TrimSpace(text), "/start") {
		b.sendMessage(ctx, message.Chat.ID, "Пришлите или перешлите сообщение со ссылкой — я сохраню его в ViArticles.")
		return
	}

	entities := append(message.Entities, message.CaptionEntities...)
	urls := extractURLs(text, entities)
	if len(urls) == 0 {
		b.sendMessage(ctx, message.Chat.ID, "Не нашёл ссылку в сообщении.")
		return
	}

	saved := 0
	for index, articleURL := range urls {
		sourceID := fmt.Sprintf("%d:%d:%d", message.Chat.ID, message.ID, index)
		created, err := b.store.CreateImportedArticle(ctx, storage.NewArticle{
			Title:    titleFor(text, articleURL),
			Body:     text,
			URL:      articleURL,
			Source:   "telegram",
			SourceID: &sourceID,
		})
		if err != nil {
			slog.Error("save telegram article", "error", err, "message_id", message.ID)
			continue
		}
		if created {
			saved++
		}
	}
	b.sendMessage(ctx, message.Chat.ID, fmt.Sprintf("Сохранено новых материалов: %d", saved))
}

func extractURLs(text string, entities []entity) []string {
	values := urlPattern.FindAllString(text, -1)
	for _, entity := range entities {
		if entity.Type == "text_link" && entity.URL != "" {
			values = append(values, entity.URL)
		}
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimRight(value, ".,!?;:)]}\"")
		parsed, err := url.ParseRequestURI(value)
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func titleFor(text, articleURL string) string {
	firstLine := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	firstLine = strings.TrimSpace(strings.ReplaceAll(firstLine, articleURL, ""))
	if firstLine != "" {
		if len([]rune(firstLine)) > 160 {
			return string([]rune(firstLine)[:160])
		}
		return firstLine
	}
	parsed, err := url.Parse(articleURL)
	if err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return "Материал из Telegram"
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string) {
	if err := b.call(ctx, "sendMessage", map[string]any{"chat_id": chatID, "text": text}, nil); err != nil && ctx.Err() == nil {
		slog.Error("send telegram response", "error", err)
	}
}

func (b *Bot) call(ctx context.Context, method string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/bot%s/%s", apiBaseURL, b.token, method), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := b.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var envelope apiResponse
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK || !envelope.OK {
		return fmt.Errorf("telegram %s: %s", method, envelope.Description)
	}
	if result != nil {
		if err := json.Unmarshal(envelope.Result, result); err != nil {
			return err
		}
	}
	return nil
}

func wait(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	Description string          `json:"description"`
}

type update struct {
	ID      int64    `json:"update_id"`
	Message *message `json:"message"`
}

type message struct {
	ID              int64    `json:"message_id"`
	Chat            chat     `json:"chat"`
	Text            string   `json:"text"`
	Caption         string   `json:"caption"`
	Entities        []entity `json:"entities"`
	CaptionEntities []entity `json:"caption_entities"`
}

type chat struct {
	ID int64 `json:"id"`
}

type entity struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
