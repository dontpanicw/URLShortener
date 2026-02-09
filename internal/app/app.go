package app

import (
	"database/sql"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/adapter/repository/postgres"
	http2 "github.com/dontpanicw/URLShortener/internal/input/http"
	"github.com/dontpanicw/URLShortener/internal/usecases"
	"github.com/dontpanicw/URLShortener/pkg/migrations"
	"log"
	"net/http"

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

	// Инициализация use cases
	urlUsecases := usecases.NewUrlUsecases(urlRepo, analyticsRepo)
	analyticsUsecases := usecases.NewAnalyticsUsecases(urlRepo, analyticsRepo)

	// Создание HTTP сервера
	srv := http2.NewServer(urlUsecases, analyticsUsecases)

	log.Printf("Starting server on %s", cfg.HTTPPort)
	return http.ListenAndServe(cfg.HTTPPort, srv)
}
