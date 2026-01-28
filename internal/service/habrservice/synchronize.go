package habrservice

import (
	"context"

	"github.com/Vikot10/viarticles/internal/dto"
)

func (h *HabrService) SynchronizeArticles(ctx context.Context) error {
	h.logger.Info().Msg("starting habr bookmarks synchronization")

	totalSynced := 0
	totalSkipped := 0
	page := 1

	for {
		articles, err := h.GetBookmarks(ctx, page)
		if err != nil {
			h.logger.Error().Err(err).Int("page", page).Msg("failed to get bookmarks from habr")
			return err
		}

		if len(articles) == 0 {
			break
		}

		for _, article := range articles {
			sourceID := "habr_" + article.ID

			exists, err := h.store.ArticleExistsBySourceID(ctx, dto.SourceHabr, sourceID)
			if err != nil {
				h.logger.Error().Err(err).Str("source_id", sourceID).Msg("failed to check article existence")
				continue
			}

			if exists {
				totalSkipped++
				continue
			}

			categories := append([]string{"habr"}, article.Hubs...)

			req := dto.CreateArticleRequest{
				Title:      article.Title,
				Body:       article.Description,
				URL:        article.URL,
				Source:     dto.SourceHabr,
				SourceID:   &sourceID,
				Categories: categories,
			}

			_, err = h.store.CreateArticle(ctx, req)
			if err != nil {
				h.logger.Error().Err(err).Str("url", article.URL).Msg("failed to create article")
				continue
			}

			totalSynced++
		}

		page++

		// Safety limit
		if page > 100 {
			h.logger.Warn().Msg("reached page limit, stopping")
			break
		}
	}

	h.logger.Info().
		Int("synced", totalSynced).
		Int("skipped", totalSkipped).
		Msg("habr bookmarks synchronization completed")

	return nil
}
