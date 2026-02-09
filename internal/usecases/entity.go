package usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"github.com/dontpanicw/URLShortener/internal/port"
	"log"
	"strings"
	"time"
)

type UrlUsecases struct {
	urlRepo       port.UrlRepository
	analyticsRepo port.AnalyticsRepository
	cache         port.CacheRepository
}

func NewUrlUsecases(urlRepo port.UrlRepository, analyticsRepo port.AnalyticsRepository, cache port.CacheRepository) *UrlUsecases {
	return &UrlUsecases{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		cache:         cache,
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

	// Сохраняем в кэш
	if u.cache != nil {
		if err := u.cache.SetUrl(ctx, &url, 1*time.Hour); err != nil {
			log.Printf("Failed to cache URL: %v", err)
		}
	}

	return &url, nil
}

func (u *UrlUsecases) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	// Сначала проверяем кэш
	if u.cache != nil {
		cachedUrl, err := u.cache.GetUrl(ctx, shortUrl)
		if err != nil {
			log.Printf("Cache error: %v", err)
		} else if cachedUrl != nil {
			log.Printf("Cache HIT for %s", shortUrl)
			return cachedUrl.FullUrl, nil
		}
		log.Printf("Cache MISS for %s", shortUrl)
	}

	// Если не в кэше, идем в БД
	fullUrl, err := u.urlRepo.GetOriginalUrl(ctx, shortUrl)
	if err != nil {
		return "", err
	}

	// Если нашли в БД, сохраняем в кэш
	if fullUrl != "" && u.cache != nil {
		urlData, err := u.urlRepo.GetUrlByShortUrl(ctx, shortUrl)
		if err == nil && urlData != nil {
			if err := u.cache.SetUrl(ctx, urlData, 1*time.Hour); err != nil {
				log.Printf("Failed to cache URL: %v", err)
			}
		}
	}

	return fullUrl, nil
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

	// Записываем клик в БД
	if err := u.analyticsRepo.RecordClick(ctx, click); err != nil {
		return err
	}

	// Увеличиваем счетчик в кэше
	if u.cache != nil {
		if _, err := u.cache.IncrementClickCount(ctx, shortUrl); err != nil {
			log.Printf("Failed to increment click count in cache: %v", err)
		}
	}

	return nil
}

func (u *UrlUsecases) GetPopularUrls(ctx context.Context, limit int) ([]domain.Url, error) {
	if u.cache == nil {
		return []domain.Url{}, nil
	}

	// Получаем список популярных shortUrl из кэша
	popularShortUrls, err := u.cache.GetPopularUrls(ctx, int64(limit))
	if err != nil {
		return nil, err
	}

	if len(popularShortUrls) == 0 {
		return []domain.Url{}, nil
	}

	// Получаем полные данные URL
	var urls []domain.Url
	for _, shortUrl := range popularShortUrls {
		urlData, err := u.urlRepo.GetUrlByShortUrl(ctx, shortUrl)
		if err != nil {
			log.Printf("Failed to get URL data for %s: %v", shortUrl, err)
			continue
		}
		if urlData != nil {
			urls = append(urls, *urlData)
		}
	}

	return urls, nil
}

func (u *UrlUsecases) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	if u.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}, nil
	}

	stats, err := u.cache.GetCacheStats(ctx)
	if err != nil {
		return nil, err
	}

	stats["enabled"] = true
	return stats, nil
}

type AnalyticsUsecases struct {
	urlRepo       port.UrlRepository
	analyticsRepo port.AnalyticsRepository
	cache         port.CacheRepository
}

func NewAnalyticsUsecases(urlRepo port.UrlRepository, analyticsRepo port.AnalyticsRepository, cache port.CacheRepository) *AnalyticsUsecases {
	return &AnalyticsUsecases{
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		cache:         cache,
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

	// Получаем счетчик из кэша для сравнения
	var cachedCount int64
	if a.cache != nil {
		cachedCount, _ = a.cache.GetClickCount(ctx, shortUrl)
	}

	return map[string]interface{}{
		"short_url":      shortUrl,
		"full_url":       urlData.FullUrl,
		"total_clicks":   totalClicks,
		"clicks_by_day":  clicksByDay,
		"cached_clicks":  cachedCount,
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
