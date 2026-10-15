package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"portal/internal/models"
)

func TestWeChatContentProviderFetchPublishedArticles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/token":
			if r.URL.Query().Get("appid") != "wx-test" || r.URL.Query().Get("secret") != "test-secret" {
				t.Fatalf("unexpected token credentials")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "token-1", "expires_in": 7200})
		case "/cgi-bin/freepublish/batchget":
			if r.URL.Query().Get("access_token") != "token-1" {
				t.Fatalf("unexpected access token")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"total_count": 1,
				"item_count":  1,
				"item": []any{map[string]any{
					"article_id":  "article-1",
					"update_time": 1_700_000_000,
					"content": map[string]any{"news_item": []any{
						map[string]any{"title": "第一篇", "author": "作者", "digest": "摘要", "content": "<p>正文</p>", "url": "https://example.com/1", "thumb_url": "https://example.com/1.jpg", "is_deleted": false},
						map[string]any{"title": "第二篇", "content": "<p>正文2</p>", "url": "https://example.com/2"},
					}},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	originalBaseURL := weChatAPIBaseURL
	weChatAPIBaseURL = server.URL
	defer func() { weChatAPIBaseURL = originalBaseURL }()
	provider := &WeChatContentProvider{}
	source := &models.ContentSource{
		ID: 991, AppID: "wx-test", AppSecret: "test-secret",
		SyncMode: models.ContentSourceSyncPublished,
	}
	provider.InvalidateToken(source.ID)
	var articles []ExternalContentArticle
	scanned, err := provider.FetchArticles(context.Background(), source, func(article ExternalContentArticle) error {
		articles = append(articles, article)
		return nil
	})
	if err != nil {
		t.Fatalf("FetchArticles() error = %v", err)
	}
	if scanned != 2 || len(articles) != 2 {
		t.Fatalf("scanned=%d len=%d, want 2", scanned, len(articles))
	}
	if articles[0].ExternalID != "article-1" || articles[0].ItemIndex != 0 || articles[0].Title != "第一篇" {
		t.Fatalf("unexpected first article: %+v", articles[0])
	}
	if articles[1].ItemIndex != 1 || articles[1].Title != "第二篇" {
		t.Fatalf("unexpected second article: %+v", articles[1])
	}
}

func TestWeChatContentProviderRequiresSavedSecret(t *testing.T) {
	provider := &WeChatContentProvider{}
	source := &models.ContentSource{ID: 992, AppID: "wx-test"}
	provider.InvalidateToken(source.ID)
	if err := provider.TestConnection(context.Background(), source); err == nil {
		t.Fatal("TestConnection() expected missing secret error")
	}
}
