package telegram

import (
	"reflect"
	"testing"
)

func TestExtractURLs(t *testing.T) {
	t.Parallel()
	got := extractURLs("Первая https://example.com/a. Повтор https://example.com/a", []entity{
		{Type: "text_link", URL: "https://go.dev/"},
	})
	want := []string{"https://example.com/a", "https://go.dev/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractURLs() = %#v, want %#v", got, want)
	}
}
