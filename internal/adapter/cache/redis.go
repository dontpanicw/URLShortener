package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"github.com/dontpanicw/URLShortener/internal/port"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	_ port.CacheRepository = (*RedisCache)(nil)
)

const (
	urlCachePrefix      = "url:"
	clickCountPrefix    = "clicks:"
	popularUrlsKey      = "popular:urls"
	defaultCacheTTL     = 1 * time.Hour
	popularUrlsCacheTTL = 5 * time.Minute
	clickThreshold      = 10 // Минимум кликов для попадания в кэш
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg *config.Config) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	return &RedisCache{
		client: client,
	}
}

// Ping проверяет соединение с Redis
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetUrl получает URL из кэша
func (r *RedisCache) GetUrl(ctx context.Context, shortUrl string) (*domain.Url, error) {
	key := urlCachePrefix + shortUrl
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Не найдено в кэше
	}
	if err != nil {
		return nil, err
	}

	var url domain.Url
	if err := json.Unmarshal([]byte(data), &url); err != nil {
		return nil, err
	}

	return &url, nil
}

// SetUrl сохраняет URL в кэш
func (r *RedisCache) SetUrl(ctx context.Context, url *domain.Url, ttl time.Duration) error {
	key := urlCachePrefix + url.ShortUrl
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	if ttl == 0 {
		ttl = defaultCacheTTL
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// DeleteUrl удаляет URL из кэша
func (r *RedisCache) DeleteUrl(ctx context.Context, shortUrl string) error {
	key := urlCachePrefix + shortUrl
	return r.client.Del(ctx, key).Err()
}

// IncrementClickCount увеличивает счетчик кликов
func (r *RedisCache) IncrementClickCount(ctx context.Context, shortUrl string) (int64, error) {
	key := clickCountPrefix + shortUrl
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Устанавливаем TTL при первом инкременте
	if count == 1 {
		r.client.Expire(ctx, key, 24*time.Hour)
	}

	// Если ссылка стала популярной, добавляем в sorted set
	if count >= clickThreshold {
		r.client.ZIncrBy(ctx, popularUrlsKey, 1, shortUrl)
		r.client.Expire(ctx, popularUrlsKey, popularUrlsCacheTTL)
	}

	return count, nil
}

// GetClickCount получает количество кликов из кэша
func (r *RedisCache) GetClickCount(ctx context.Context, shortUrl string) (int64, error) {
	key := clickCountPrefix + shortUrl
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetPopularUrls получает список популярных URL
func (r *RedisCache) GetPopularUrls(ctx context.Context, limit int64) ([]string, error) {
	// Получаем топ URL по количеству кликов (в обратном порядке)
	urls, err := r.client.ZRevRange(ctx, popularUrlsKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	return urls, nil
}

// InvalidateUrl инвалидирует кэш для конкретного URL
func (r *RedisCache) InvalidateUrl(ctx context.Context, shortUrl string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, urlCachePrefix+shortUrl)
	pipe.Del(ctx, clickCountPrefix+shortUrl)
	pipe.ZRem(ctx, popularUrlsKey, shortUrl)
	_, err := pipe.Exec(ctx)
	return err
}

// WarmupCache предзагружает популярные URL в кэш
func (r *RedisCache) WarmupCache(ctx context.Context, urls []domain.Url) error {
	if len(urls) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for _, url := range urls {
		key := urlCachePrefix + url.ShortUrl
		data, err := json.Marshal(url)
		if err != nil {
			continue
		}
		pipe.Set(ctx, key, data, defaultCacheTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// GetCacheStats возвращает статистику кэша
func (r *RedisCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}

	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}

	popularCount, err := r.client.ZCard(ctx, popularUrlsKey).Result()
	if err != nil {
		popularCount = 0
	}

	return map[string]interface{}{
		"db_size":       dbSize,
		"popular_count": popularCount,
		"info":          info,
	}, nil
}

// Close закрывает соединение с Redis
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// BatchGetUrls получает несколько URL за один запрос
func (r *RedisCache) BatchGetUrls(ctx context.Context, shortUrls []string) (map[string]*domain.Url, error) {
	if len(shortUrls) == 0 {
		return make(map[string]*domain.Url), nil
	}

	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, shortUrl := range shortUrls {
		key := urlCachePrefix + shortUrl
		cmds[shortUrl] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]*domain.Url)
	for shortUrl, cmd := range cmds {
		data, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue
		}

		var url domain.Url
		if err := json.Unmarshal([]byte(data), &url); err != nil {
			continue
		}
		result[shortUrl] = &url
	}

	return result, nil
}

// SetClickCountBatch устанавливает счетчики кликов пакетом
func (r *RedisCache) SetClickCountBatch(ctx context.Context, counts map[string]int64) error {
	if len(counts) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for shortUrl, count := range counts {
		key := clickCountPrefix + shortUrl
		pipe.Set(ctx, key, count, 24*time.Hour)

		if count >= clickThreshold {
			pipe.ZAdd(ctx, popularUrlsKey, redis.Z{
				Score:  float64(count),
				Member: shortUrl,
			})
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Helper function для создания ключа
func (r *RedisCache) makeKey(prefix, value string) string {
	return fmt.Sprintf("%s%s", prefix, value)
}
