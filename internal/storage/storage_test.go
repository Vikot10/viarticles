package storage

import (
	"context"
	"testing"
)

func TestArticleLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	article, err := store.CreateArticle(ctx, NewArticle{
		Title: "Go", URL: "https://go.dev", Source: "manual",
	})
	if err != nil {
		t.Fatal(err)
	}
	if article.Read || article.Liked {
		t.Fatalf("new article has unexpected flags: %#v", article)
	}

	article, err = store.ToggleRead(ctx, article.ID)
	if err != nil || !article.Read {
		t.Fatalf("ToggleRead() article=%#v err=%v", article, err)
	}
	article, err = store.ToggleLiked(ctx, article.ID)
	if err != nil || !article.Liked {
		t.Fatalf("ToggleLiked() article=%#v err=%v", article, err)
	}

	liked, err := store.ListArticles(ctx, "liked", "")
	if err != nil || len(liked) != 1 {
		t.Fatalf("ListArticles(liked) len=%d err=%v", len(liked), err)
	}
	if err := store.DeleteArticle(ctx, article.ID); err != nil {
		t.Fatal(err)
	}
	articles, err := store.ListArticles(ctx, "", "")
	if err != nil || len(articles) != 0 {
		t.Fatalf("ListArticles() after delete len=%d err=%v", len(articles), err)
	}
}

func TestImportedArticleIsIdempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	sourceID := "1:2:0"
	input := NewArticle{
		Title: "Telegram", URL: "https://example.com", Source: "telegram", SourceID: &sourceID,
	}
	created, err := store.CreateImportedArticle(ctx, input)
	if err != nil || !created {
		t.Fatalf("first import created=%v err=%v", created, err)
	}
	created, err = store.CreateImportedArticle(ctx, input)
	if err != nil || created {
		t.Fatalf("second import created=%v err=%v", created, err)
	}
}
