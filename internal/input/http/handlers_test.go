package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockUrlUsecases struct{}

func (m *MockUrlUsecases) CreateShortUrl(ctx context.Context, fullUrl string, customShort string) (*domain.Url, error) {
	shortUrl := customShort
	if shortUrl == "" {
		shortUrl = "abc123"
	}
	return &domain.Url{
		Id:        "test-id",
		FullUrl:   fullUrl,
		ShortUrl:  shortUrl,
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockUrlUsecases) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	if shortUrl == "test" {
		return "https://example.com", nil
	}
	return "", nil
}

func (m *MockUrlUsecases) RecordClick(ctx context.Context, shortUrl string, ipAddress string, userAgent string) error {
	return nil
}

func (m *MockUrlUsecases) GetPopularUrls(ctx context.Context, limit int) ([]domain.Url, error) {
	return []domain.Url{
		{
			Id:        "1",
			FullUrl:   "https://example.com",
			ShortUrl:  "test",
			CreatedAt: time.Now(),
		},
	}, nil
}

func (m *MockUrlUsecases) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"enabled": true,
		"db_size": 10,
	}, nil
}

type MockAnalyticsUsecases struct{}

func (m *MockAnalyticsUsecases) GetClickStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return map[string]interface{}{
		"short_url":    shortUrl,
		"full_url":     "https://example.com",
		"total_clicks": 42,
	}, nil
}

func (m *MockAnalyticsUsecases) GetUserAgentStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return []map[string]interface{}{
		{
			"user_agent": "Mozilla/5.0",
			"count":      25,
		},
	}, nil
}

func TestShortenUrl(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Create short URL successfully", func(t *testing.T) {
		reqBody := ShortenRequest{
			Url:         "https://example.com",
			CustomShort: "test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ShortenUrl(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}

		var response ShortenResponse
		json.NewDecoder(w.Body).Decode(&response)

		if response.ShortUrl != "test" {
			t.Errorf("Expected short_url 'test', got '%s'", response.ShortUrl)
		}
	})

	t.Run("Invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer([]byte("invalid")))
		w := httptest.NewRecorder()

		handler.ShortenUrl(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Empty URL", func(t *testing.T) {
		reqBody := ShortenRequest{Url: ""}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.ShortenUrl(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestRedirectUrl(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Redirect existing URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.RedirectUrl(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status %d, got %d", http.StatusMovedPermanently, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "https://example.com" {
			t.Errorf("Expected location 'https://example.com', got '%s'", location)
		}
	})

	t.Run("Non-existing URL", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.RedirectUrl(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestGetAnalytics(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Get analytics successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/analytics/test", nil)
		w := httptest.NewRecorder()

		handler.GetAnalytics(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		if response["short_url"] != "test" {
			t.Errorf("Expected short_url 'test', got '%v'", response["short_url"])
		}
	})

	t.Run("Get analytics with date filters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/analytics/test?start_date=2026-01-01&end_date=2026-02-09", nil)
		w := httptest.NewRecorder()

		handler.GetAnalytics(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestGetUserAgentStats(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Get user agent stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/analytics/test/user-agents", nil)
		w := httptest.NewRecorder()

		handler.GetUserAgentStats(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		stats, ok := response["user_agent_stats"].([]interface{})
		if !ok || len(stats) == 0 {
			t.Error("Expected user_agent_stats to be non-empty array")
		}
	})
}

func TestGetPopularUrls(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Get popular URLs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/popular", nil)
		w := httptest.NewRecorder()

		handler.GetPopularUrls(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		count, ok := response["count"].(float64)
		if !ok || count != 1 {
			t.Errorf("Expected count 1, got %v", count)
		}
	})
}

func TestGetCacheStats(t *testing.T) {
	handler := NewHandler(&MockUrlUsecases{}, &MockAnalyticsUsecases{})

	t.Run("Get cache stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cache/stats", nil)
		w := httptest.NewRecorder()

		handler.GetCacheStats(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		enabled, ok := response["enabled"].(bool)
		if !ok || !enabled {
			t.Error("Expected cache to be enabled")
		}
	})
}
