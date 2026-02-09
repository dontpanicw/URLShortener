package postgres

import (
	"context"
	"database/sql"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"github.com/dontpanicw/URLShortener/internal/port"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"time"
)

var (
	_ port.UrlRepository = (*UrlRepository)(nil)
)

const (
	createUrlQuery      = `INSERT INTO urls (id, full_url, short_url, created_at) VALUES ($1, $2, $3, $4)`
	getOriginalUrlQuery = `SELECT full_url FROM urls WHERE short_url = $1`
	getUrlByShortQuery  = `SELECT id, full_url, short_url, created_at FROM urls WHERE short_url = $1`
	checkExistsQuery    = `SELECT EXISTS(SELECT 1 FROM urls WHERE short_url = $1)`
)

type UrlRepository struct {
	PostgresDB *dbpg.DB
}

func NewUrlRepository(config *config.Config) *UrlRepository {
	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(config.MasterDSN, config.SlaveDSNs, opts)
	if err != nil {
		panic(err)
	}

	return &UrlRepository{
		PostgresDB: db,
	}
}

func (u *UrlRepository) CreateShortUrl(ctx context.Context, url domain.Url) error {
	_, err := u.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), createUrlQuery, url.Id, url.FullUrl, url.ShortUrl, url.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (u *UrlRepository) GetOriginalUrl(ctx context.Context, shortUrl string) (string, error) {
	var fullUrl string
	err := u.PostgresDB.QueryRowContext(ctx, getOriginalUrlQuery, shortUrl).Scan(&fullUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return fullUrl, nil
}

func (u *UrlRepository) GetUrlByShortUrl(ctx context.Context, shortUrl string) (*domain.Url, error) {
	var url domain.Url
	err := u.PostgresDB.QueryRowContext(ctx, getUrlByShortQuery, shortUrl).Scan(&url.Id, &url.FullUrl, &url.ShortUrl, &url.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (u *UrlRepository) ShortUrlExists(ctx context.Context, shortUrl string) (bool, error) {
	var exists bool
	err := u.PostgresDB.QueryRowContext(ctx, checkExistsQuery, shortUrl).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func createRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: 3,
		Delay:    5 * time.Second,
		Backoff:  2}
}
