package webapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Vikot10/viarticles/internal/storage"
)

func TestManualArticleFlow(t *testing.T) {
	t.Parallel()
	store, err := storage.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	server, err := New(store)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{"title": {"Go"}, "url": {"https://go.dev"}}
	request := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("POST /articles status=%d body=%s", response.Code, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "https://go.dev") {
		t.Fatalf("GET / status=%d body=%s", response.Code, response.Body.String())
	}
}
