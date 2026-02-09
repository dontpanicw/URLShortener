package usecases

import (
	"context"
	"errors"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"testing"
	"time"
)

// Mock репозиториев
type MockUrlRepository struct {
	urls      map[string]*domain.Url
	shortUrls map[string]string
}

func NewMockUrlRepository() *MockUrlRepository {
	return &MockUrlRepository{
		urls:      make(map[string]*domain.Url),
		shortUrls: make(map[string]string),
	}
}

func (m *MockUrlRepository) CreateShortUrl(ctx context.Context, url domain.Url) error {
	if _, exists := m.shortUrls[url.ShortUrl]; exists {
		return errors.New("short URL already exists")
	}
	m.urls[url.Id] = &url
	m.shortUrls[url.ShortUrl] = url.Id
	return nil
}

func (m *MockUrlRepository) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	id, exists := m.shortUrls[shortUrl]
	if !exists {
		return "", nil
	}
	url := m.urls[id]
	return url.FullUrl, nil
}

func (m *MockUrlRepository) GetUrlByShortUrl(ctx context.Context, shortUrl string) (*domain.Url, error) {
	id, exists := m.shortUrls[shortUrl]
	if !exists {
		return nil, errors.New("URL not found")
	}
	return m.urls[id], nil
}

func (m *MockUrlRepository) ShortUrlExists(ctx context.Context, shortUrl string) (bool, error) {
	_, exists := m.shortUrls[shortUrl]
	return exists, nil
}

type MockAnalyticsRepository struct {
	clicks []domain.URLClick
}

func NewMockAnalyticsRepository() *MockAnalyticsRepository {
	return &MockAnalyticsRepository{
		clicks: make([]domain.URLClick, 0),
	}
}

func (m *MockAnalyticsRepository) RecordClick(ctx context.Context, click domain.URLClick) error {
	m.clicks = append(m.clicks, click)
	return nil
}

func (m *MockAnalyticsRepository) GetClickCount(ctx context.Context, shortUrlID string, startDate, endDate time.Time) (int, error) {
	count := 0
	for _, click := range m.clicks {
		if click.ShortURLID == shortUrlID && click.ClickedAt.After(startDate) && click.ClickedAt.Before(endDate) {
			count++
		}
	}
	return count, nil
}

func (m *MockAnalyticsRepository) GetClicksByPeriod(ctx context.Context, shortUrlID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

func (m *MockAnalyticsRepository) GetUserAgentStats(ctx context.Context, shortUrlID string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return []map[string]interface{}{}, nil
}

type MockCacheRepository struct {
	urls        map[string]*domain.Url
	clickCounts map[string]int64
	popular     []string
}

func NewMockCacheRepository() *MockCacheRepository {
	return &MockCacheRepository{
		urls:        make(map[string]*domain.Url),
		clickCounts: make(map[string]int64),
		popular:     make([]string, 0),
	}
}

func (m *MockCacheRepository) GetUrl(ctx context.Context, shortUrl string) (*domain.Url, error) {
	url, exists := m.urls[shortUrl]
	if !exists {
		return nil, nil
	}
	return url, nil
}

func (m *MockCacheRepository) SetUrl(ctx context.Context, url *domain.Url, ttl time.Duration) error {
	m.urls[url.ShortUrl] = url
	return nil
}

func (m *MockCacheRepository) DeleteUrl(ctx context.Context, shortUrl string) error {
	delete(m.urls, shortUrl)
	return nil
}

func (m *MockCacheRepository) BatchGetUrls(ctx context.Context, shortUrls []string) (map[string]*domain.Url, error) {
	result := make(map[string]*domain.Url)
	for _, shortUrl := range shortUrls {
		if url, exists := m.urls[shortUrl]; exists {
			result[shortUrl] = url
		}
	}
	return result, nil
}

func (m *MockCacheRepository) IncrementClickCount(ctx context.Context, shortUrl string) (int64, error) {
	m.clickCounts[shortUrl]++
	count := m.clickCounts[shortUrl]
	if count >= 10 {
		m.popular = append(m.popular, shortUrl)
	}
	return count, nil
}

func (m *MockCacheRepository) GetClickCount(ctx context.Context, shortUrl string) (int64, error) {
	return m.clickCounts[shortUrl], nil
}

func (m *MockCacheRepository) SetClickCountBatch(ctx context.Context, counts map[string]int64) error {
	for shortUrl, count := range counts {
		m.clickCounts[shortUrl] = count
	}
	return nil
}

func (m *MockCacheRepository) GetPopularUrls(ctx context.Context, limit int64) ([]string, error) {
	if int64(len(m.popular)) > limit {
		return m.popular[:limit], nil
	}
	return m.popular, nil
}

func (m *MockCacheRepository) InvalidateUrl(ctx context.Context, shortUrl string) error {
	delete(m.urls, shortUrl)
	delete(m.clickCounts, shortUrl)
	return nil
}

func (m *MockCacheRepository) WarmupCache(ctx context.Context, urls []domain.Url) error {
	for _, url := range urls {
		m.urls[url.ShortUrl] = &url
	}
	return nil
}

func (m *MockCacheRepository) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"db_size":       len(m.urls),
		"popular_count": len(m.popular),
	}, nil
}

func (m *MockCacheRepository) Ping(ctx context.Context) error {
	return nil
}

func (m *MockCacheRepository) Close() error {
	return nil
}

// Тесты
func TestCreateShortUrl(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	uc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	t.Run("Create with auto-generated short URL", func(t *testing.T) {
		url, err := uc.CreateShortUrl(ctx, "https://example.com", "")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if url.FullUrl != "https://example.com" {
			t.Errorf("Expected full URL 'https://example.com', got '%s'", url.FullUrl)
		}
		if url.ShortUrl == "" {
			t.Error("Expected short URL to be generated")
		}
	})

	t.Run("Create with custom short URL", func(t *testing.T) {
		url, err := uc.CreateShortUrl(ctx, "https://github.com", "gh")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if url.ShortUrl != "gh" {
			t.Errorf("Expected short URL 'gh', got '%s'", url.ShortUrl)
		}
	})

	t.Run("Create with duplicate custom short URL", func(t *testing.T) {
		_, err := uc.CreateShortUrl(ctx, "https://google.com", "gh")
		if err == nil {
			t.Error("Expected error for duplicate short URL")
		}
	})

	t.Run("Create with empty URL", func(t *testing.T) {
		_, err := uc.CreateShortUrl(ctx, "", "test")
		if err == nil {
			t.Error("Expected error for empty URL")
		}
	})
}

func TestGetOriginalUrl(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	uc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	// Создаем тестовую ссылку
	testUrl, _ := uc.CreateShortUrl(ctx, "https://example.com", "test")

	t.Run("Get existing URL from cache", func(t *testing.T) {
		fullUrl, err := uc.GetOriginalUrl(ctx, testUrl.ShortUrl)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fullUrl != "https://example.com" {
			t.Errorf("Expected 'https://example.com', got '%s'", fullUrl)
		}
	})

	t.Run("Get non-existing URL", func(t *testing.T) {
		fullUrl, err := uc.GetOriginalUrl(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fullUrl != "" {
			t.Errorf("Expected empty string, got '%s'", fullUrl)
		}
	})
}

func TestRecordClick(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	uc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	// Создаем тестовую ссылку
	testUrl, _ := uc.CreateShortUrl(ctx, "https://example.com", "test")

	t.Run("Record click successfully", func(t *testing.T) {
		err := uc.RecordClick(ctx, testUrl.ShortUrl, "127.0.0.1", "Mozilla/5.0")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Проверяем, что клик записан
		if len(analyticsRepo.clicks) != 1 {
			t.Errorf("Expected 1 click, got %d", len(analyticsRepo.clicks))
		}

		// Проверяем счетчик в кэше
		count, _ := cache.GetClickCount(ctx, testUrl.ShortUrl)
		if count != 1 {
			t.Errorf("Expected click count 1, got %d", count)
		}
	})

	t.Run("Record multiple clicks", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			uc.RecordClick(ctx, testUrl.ShortUrl, "127.0.0.1", "Mozilla/5.0")
		}

		count, _ := cache.GetClickCount(ctx, testUrl.ShortUrl)
		if count != 6 { // 1 from previous test + 5 new
			t.Errorf("Expected click count 6, got %d", count)
		}
	})
}

func TestGetPopularUrls(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	uc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	// Создаем несколько ссылок
	url1, _ := uc.CreateShortUrl(ctx, "https://example1.com", "test1")
	url2, _ := uc.CreateShortUrl(ctx, "https://example2.com", "test2")

	// Делаем много кликов для первой ссылки
	for i := 0; i < 15; i++ {
		uc.RecordClick(ctx, url1.ShortUrl, "127.0.0.1", "Mozilla/5.0")
	}

	// Делаем мало кликов для второй ссылки
	for i := 0; i < 5; i++ {
		uc.RecordClick(ctx, url2.ShortUrl, "127.0.0.1", "Mozilla/5.0")
	}

	t.Run("Get popular URLs", func(t *testing.T) {
		popular, err := uc.GetPopularUrls(ctx, 10)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Первая ссылка должна быть в популярных (>= 10 кликов)
		found := false
		for _, url := range popular {
			if url.ShortUrl == url1.ShortUrl {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected test1 to be in popular URLs")
		}
	})
}

func TestGetCacheStats(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	uc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	// Создаем несколько ссылок
	uc.CreateShortUrl(ctx, "https://example1.com", "test1")
	uc.CreateShortUrl(ctx, "https://example2.com", "test2")

	t.Run("Get cache stats", func(t *testing.T) {
		stats, err := uc.GetCacheStats(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		enabled, ok := stats["enabled"].(bool)
		if !ok || !enabled {
			t.Error("Expected cache to be enabled")
		}

		dbSize, ok := stats["db_size"].(int)
		if !ok || dbSize != 2 {
			t.Errorf("Expected db_size 2, got %v", dbSize)
		}
	})
}

func TestAnalyticsUsecases(t *testing.T) {
	urlRepo := NewMockUrlRepository()
	analyticsRepo := NewMockAnalyticsRepository()
	cache := NewMockCacheRepository()
	
	urlUc := NewUrlUsecases(urlRepo, analyticsRepo, cache)
	analyticsUc := NewAnalyticsUsecases(urlRepo, analyticsRepo, cache)
	ctx := context.Background()

	// Создаем тестовую ссылку и клики
	testUrl, _ := urlUc.CreateShortUrl(ctx, "https://example.com", "test")
	for i := 0; i < 5; i++ {
		urlUc.RecordClick(ctx, testUrl.ShortUrl, "127.0.0.1", "Mozilla/5.0")
	}

	t.Run("Get click stats", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now().Add(24 * time.Hour)

		stats, err := analyticsUc.GetClickStats(ctx, testUrl.ShortUrl, startDate, endDate)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		shortUrl, ok := stats["short_url"].(string)
		if !ok || shortUrl != "test" {
			t.Errorf("Expected short_url 'test', got %v", shortUrl)
		}

		fullUrl, ok := stats["full_url"].(string)
		if !ok || fullUrl != "https://example.com" {
			t.Errorf("Expected full_url 'https://example.com', got %v", fullUrl)
		}
	})
}

func TestGenerateShortUrl(t *testing.T) {
	t.Run("Generate short URL", func(t *testing.T) {
		shortUrl, err := generateShortUrl()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(shortUrl) != 6 {
			t.Errorf("Expected length 6, got %d", len(shortUrl))
		}
	})

	t.Run("Generate unique short URLs", func(t *testing.T) {
		urls := make(map[string]bool)
		for i := 0; i < 100; i++ {
			shortUrl, _ := generateShortUrl()
			if urls[shortUrl] {
				t.Errorf("Generated duplicate short URL: %s", shortUrl)
			}
			urls[shortUrl] = true
		}
	})
}

func TestGenerateID(t *testing.T) {
	t.Run("Generate ID", func(t *testing.T) {
		id := generateID()
		if id == "" {
			t.Error("Expected non-empty ID")
		}
	})

	t.Run("Generate unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := generateID()
			if ids[id] {
				t.Errorf("Generated duplicate ID: %s", id)
			}
			ids[id] = true
		}
	})
}
