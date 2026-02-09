package app

import (
	"context"
	"database/sql"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/adapter/cache"
	"github.com/dontpanicw/URLShortener/internal/adapter/repository/postgres"
	http2 "github.com/dontpanicw/URLShortener/internal/input/http"
	"github.com/dontpanicw/URLShortener/internal/usecases"
	"github.com/dontpanicw/URLShortener/pkg/migrations"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func Start(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.MasterDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrations.Migrate(db); err != nil {
		return err
	}
	log.Print("Migrations applied successfully")

	// Инициализация репозиториев
	urlRepo := postgres.NewUrlRepository(cfg)
	analyticsRepo := postgres.NewAnalyticsRepository(cfg)

	// Инициализация Redis кэша
	redisCache := cache.NewRedisCache(cfg)
	
	// Проверяем соединение с Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := redisCache.Ping(ctx); err != nil {
		log.Printf("WARNING: Redis is not available: %v. Running without cache.", err)
		redisCache = nil // Отключаем кэш, если Redis недоступен
	} else {
		log.Print("Redis connected successfully")
		defer redisCache.Close()
	}

	// Инициализация use cases с кэшем
	urlUsecases := usecases.NewUrlUsecases(urlRepo, analyticsRepo, redisCache)
	analyticsUsecases := usecases.NewAnalyticsUsecases(urlRepo, analyticsRepo, redisCache)

	// Создание HTTP сервера
	srv := http2.NewServer(urlUsecases, analyticsUsecases)

	log.Printf("Starting server on %s", cfg.HTTPPort)
	return http.ListenAndServe(cfg.HTTPPort, srv)
}
