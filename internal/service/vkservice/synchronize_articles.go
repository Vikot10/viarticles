package vkservice

import (
	"context"

	"github.com/Vikot10/viarticles/internal/dto"
)

func (v *VkService) SynchronizeArticles(ctx context.Context) error {
	v.logger.Info().Msg("starting vk faves synchronization")

	const batchSize = 50
	offset := 0
	totalSynced := 0
	totalSkipped := 0

	for {
		faves, total, err := v.GetFaves(ctx, offset, batchSize)
		if err != nil {
			v.logger.Error().Err(err).Msg("failed to get faves from vk")
			return err
		}

		if len(faves) == 0 {
			break
		}

		for _, fave := range faves {
			if fave.URL == "" {
				continue
			}

			exists, err := v.store.ArticleExistsBySourceID(ctx, dto.SourceVK, fave.SourceID)
			if err != nil {
				v.logger.Error().Err(err).Str("source_id", fave.SourceID).Msg("failed to check article existence")
				continue
			}

			if exists {
				totalSkipped++
				continue
			}

			title := fave.Title
			if title == "" {
				title = fave.URL
			}

			req := dto.CreateArticleRequest{
				Title:      title,
				Body:       fave.Body,
				URL:        fave.URL,
				Source:     dto.SourceVK,
				SourceID:   &fave.SourceID,
				Categories: []string{"vk", fave.Type},
			}

			_, err = v.store.CreateArticle(ctx, req)
			if err != nil {
				v.logger.Error().Err(err).Str("url", fave.URL).Msg("failed to create article")
				continue
			}

			totalSynced++
		}

		offset += batchSize
		if offset >= total {
			break
		}
	}

	v.logger.Info().
		Int("synced", totalSynced).
		Int("skipped", totalSkipped).
		Msg("vk faves synchronization completed")

	return nil
}
