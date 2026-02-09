package port

import (
	"context"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"time"
)

type UrlRepository interface {
	CreateShortUrl(ctx context.Context, url domain.Url) error
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, error)
	GetUrlByShortUrl(ctx context.Context, shortUrl string) (*domain.Url, error)
	ShortUrlExists(ctx context.Context, shortUrl string) (bool, error)
}

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, click domain.URLClick) error
	GetClickCount(ctx context.Context, shortUrlID string, startDate, endDate time.Time) (int, error)
	GetClicksByPeriod(ctx context.Context, shortUrlID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error)
	GetUserAgentStats(ctx context.Context, shortUrlID string, startDate, endDate time.Time) ([]map[string]interface{}, error)
}

type CacheRepository interface {
	// URL operations
	GetUrl(ctx context.Context, shortUrl string) (*domain.Url, error)
	SetUrl(ctx context.Context, url *domain.Url, ttl time.Duration) error
	DeleteUrl(ctx context.Context, shortUrl string) error
	BatchGetUrls(ctx context.Context, shortUrls []string) (map[string]*domain.Url, error)

	// Click count operations
	IncrementClickCount(ctx context.Context, shortUrl string) (int64, error)
	GetClickCount(ctx context.Context, shortUrl string) (int64, error)
	SetClickCountBatch(ctx context.Context, counts map[string]int64) error

	// Popular URLs
	GetPopularUrls(ctx context.Context, limit int64) ([]string, error)

	// Cache management
	InvalidateUrl(ctx context.Context, shortUrl string) error
	WarmupCache(ctx context.Context, urls []domain.Url) error
	GetCacheStats(ctx context.Context) (map[string]interface{}, error)
	Ping(ctx context.Context) error
	Close() error
}
