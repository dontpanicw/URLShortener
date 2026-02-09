package port

import (
	"context"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"time"
)

type UrlUsecases interface {
	CreateShortUrl(ctx context.Context, fullUrl string, customShort string) (*domain.Url, error)
	GetOriginalUrl(ctx context.Context, shortUrl string) (string, error)
	RecordClick(ctx context.Context, shortUrl string, ipAddress string, userAgent string) error
}

type AnalyticsUsecases interface {
	GetClickStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetUserAgentStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) ([]map[string]interface{}, error)
}
