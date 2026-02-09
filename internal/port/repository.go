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
