package usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"github.com/dontpanicw/URLShortener/internal/port"
	"strings"
	"time"
)

type UrlUsecases struct {
	urlRepo       port.UrlRepository
	analyticsRepo port.AnalyticsRepository
}

func NewUrlUsecases(urlRepo port.UrlRepository, analyticsRepo port.AnalyticsRepository) *UrlUsecases {
	return &UrlUsecases{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
	}
}

func (u *UrlUsecases) CreateShortUrl(ctx context.Context, fullUrl string, customShort string) (*domain.Url, error) {
	if fullUrl == "" {
		return nil, errors.New("full URL cannot be empty")
	}

	var shortUrl string
	if customShort != "" {
		// Проверяем, не занят ли кастомный короткий URL
		exists, err := u.urlRepo.ShortUrlExists(ctx, customShort)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("custom short URL already exists")
		}
		shortUrl = customShort
	} else {
		// Генерируем случайный короткий URL
		var err error
		shortUrl, err = generateShortUrl()
		if err != nil {
			return nil, err
		}
	}

	url := domain.Url{
		Id:        generateID(),
		FullUrl:   fullUrl,
		ShortUrl:  shortUrl,
		CreatedAt: time.Now(),
	}

	if err := u.urlRepo.CreateShortUrl(ctx, url); err != nil {
		return nil, err
	}

	return &url, nil
}

func (u *UrlUsecases) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	return u.urlRepo.GetOriginalUrl(ctx, shortUrl)
}

func (u *UrlUsecases) RecordClick(ctx context.Context, shortUrl string, ipAddress string, userAgent string) error {
	urlData, err := u.urlRepo.GetUrlByShortUrl(ctx, shortUrl)
	if err != nil {
		return err
	}

	click := domain.URLClick{
		Id:         generateID(),
		ShortURLID: urlData.Id,
		ClickedAt:  time.Now(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	return u.analyticsRepo.RecordClick(ctx, click)
}

type AnalyticsUsecases struct {
	urlRepo       port.UrlRepository
	analyticsRepo port.AnalyticsRepository
}

func NewAnalyticsUsecases(urlRepo port.UrlRepository, analyticsRepo port.AnalyticsRepository) *AnalyticsUsecases {
	return &AnalyticsUsecases{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
	}
}

func (a *AnalyticsUsecases) GetClickStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) (map[string]interface{}, error) {
	urlData, err := a.urlRepo.GetUrlByShortUrl(ctx, shortUrl)
	if err != nil {
		return nil, err
	}

	totalClicks, err := a.analyticsRepo.GetClickCount(ctx, urlData.Id, startDate, endDate)
	if err != nil {
		return nil, err
	}

	clicksByDay, err := a.analyticsRepo.GetClicksByPeriod(ctx, urlData.Id, startDate, endDate, "day")
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"short_url":    shortUrl,
		"full_url":     urlData.FullUrl,
		"total_clicks": totalClicks,
		"clicks_by_day": clicksByDay,
	}, nil
}

func (a *AnalyticsUsecases) GetUserAgentStats(ctx context.Context, shortUrl string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	urlData, err := a.urlRepo.GetUrlByShortUrl(ctx, shortUrl)
	if err != nil {
		return nil, err
	}

	return a.analyticsRepo.GetUserAgentStats(ctx, urlData.Id, startDate, endDate)
}

// Вспомогательные функции
func generateShortUrl() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	return strings.TrimRight(encoded, "=")[:6], nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
