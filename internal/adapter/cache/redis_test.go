package cache

import (
	"context"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"testing"
	"time"
)

func TestRedisCache_SetAndGetUrl(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1, // Use test DB
	}

	cache := NewRedisCache(cfg)
	ctx := context.Background()

	// Skip test if Redis is not available
	if err := cache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer cache.Close()

	t.Run("Set and get URL", func(t *testing.T) {
		url := &domain.Url{
			Id:        "test-id",
			FullUrl:   "https://example.com",
			ShortUrl:  "test",
			CreatedAt: time.Now(),
		}

		err := cache.SetUrl(ctx, url, 1*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set URL: %v", err)
		}

		retrieved, err := cache.GetUrl(ctx, "test")
		if err != nil {
			t.Fatalf("Failed to get URL: %v", err)
		}

		if retrieved == nil {
			t.Fatal("Expected URL, got nil")
		}

		if retrieved.FullUrl != url.FullUrl {
			t.Errorf("Expected full URL '%s', got '%s'", url.FullUrl, retrieved.FullUrl)
		}
	})

	t.Run("Get non-existing URL", func(t *testing.T) {
		retrieved, err := cache.GetUrl(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if retrieved != nil {
			t.Error("Expected nil for non-existing URL")
		}
	})
}

func TestRedisCache_ClickCount(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1,
	}

	cache := NewRedisCache(cfg)
	ctx := context.Background()

	if err := cache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer cache.Close()

	t.Run("Increment click count", func(t *testing.T) {
		shortUrl := "test-clicks"

		count1, err := cache.IncrementClickCount(ctx, shortUrl)
		if err != nil {
			t.Fatalf("Failed to increment: %v", err)
		}
		if count1 != 1 {
			t.Errorf("Expected count 1, got %d", count1)
		}

		count2, err := cache.IncrementClickCount(ctx, shortUrl)
		if err != nil {
			t.Fatalf("Failed to increment: %v", err)
		}
		if count2 != 2 {
			t.Errorf("Expected count 2, got %d", count2)
		}
	})

	t.Run("Get click count", func(t *testing.T) {
		shortUrl := "test-get-clicks"
		cache.IncrementClickCount(ctx, shortUrl)
		cache.IncrementClickCount(ctx, shortUrl)

		count, err := cache.GetClickCount(ctx, shortUrl)
		if err != nil {
			t.Fatalf("Failed to get count: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2, got %d", count)
		}
	})
}

func TestRedisCache_PopularUrls(t *testing.T) {
	cfg := &config.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1,
	}

	cache := NewRedisCache(cfg)
	ctx := context.Background()

	if err := cache.Ping(ctx); err != nil {
		t.Skip("Redis not available, skipping test")
	}
	defer cache.Close()

	t.Run("Popular URLs after threshold", func(t *testing.T) {
		shortUrl := "popular-test"

		// Increment to reach threshold
		for i := 0; i < 15; i++ {
			cache.IncrementClickCount(ctx, shortUrl)
		}

		popular, err := cache.GetPopularUrls(ctx, 10)
		if err != nil {
			t.Fatalf("Failed to get popular URLs: %v", err)
		}

		found := false
		for _, url := range popular {
			if url == shortUrl {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected URL to be in popular list")
		}
	})
}
