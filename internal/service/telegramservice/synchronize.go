package telegramservice

import (
	"context"
	"fmt"

	"github.com/Vikot10/viarticles/internal/dto"
)

func (t *TelegramService) SynchronizeArticles(ctx context.Context) error {
	t.logger.Info().Msg("starting telegram synchronization")

	var lastOffset int64 = 0
	totalSynced := 0
	totalSkipped := 0

	for {
		updates, err := t.GetUpdates(ctx, lastOffset, 100)
		if err != nil {
			t.logger.Error().Err(err).Msg("failed to get updates from telegram")
			return err
		}

		if len(updates) == 0 {
			break
		}

		messages := t.ParseMessages(updates)

		for _, msg := range messages {
			for _, url := range msg.URLs {
				sourceID := fmt.Sprintf("tg_%d_%d", msg.ChatID, msg.MessageID)

				exists, err := t.store.ArticleExistsBySourceID(ctx, dto.SourceTelegram, sourceID)
				if err != nil {
					t.logger.Error().Err(err).Str("source_id", sourceID).Msg("failed to check article existence")
					continue
				}

				if exists {
					totalSkipped++
					continue
				}

				title := t.ExtractTitle(msg.Text, url)

				req := dto.CreateArticleRequest{
					Title:      title,
					Body:       msg.Text,
					URL:        url,
					Source:     dto.SourceTelegram,
					SourceID:   &sourceID,
					Categories: []string{"telegram"},
				}

				_, err = t.store.CreateArticle(ctx, req)
				if err != nil {
					t.logger.Error().Err(err).Str("url", url).Msg("failed to create article")
					continue
				}

				totalSynced++
			}
		}

		lastOffset = updates[len(updates)-1].UpdateID + 1
	}

	t.logger.Info().
		Int("synced", totalSynced).
		Int("skipped", totalSkipped).
		Msg("telegram synchronization completed")

	return nil
}
