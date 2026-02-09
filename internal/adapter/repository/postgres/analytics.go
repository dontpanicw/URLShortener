package postgres

import (
	"context"
	"github.com/dontpanicw/URLShortener/config"
	"github.com/dontpanicw/URLShortener/internal/domain"
	"github.com/dontpanicw/URLShortener/internal/port"
	"github.com/wb-go/wbf/dbpg"
	"time"
)

var (
	_ port.AnalyticsRepository = (*AnalyticsRepository)(nil)
)

const (
	recordClickQuery = `INSERT INTO url_clicks (short_url_id, clicked_at, ip_address, user_agent) VALUES ($1, $2, $3, $4)`
	
	getClickCountQuery = `
		SELECT COUNT(*) 
		FROM url_clicks 
		WHERE short_url_id = $1 
		AND clicked_at BETWEEN $2 AND $3`
	
	getClicksByDayQuery = `
		SELECT DATE(clicked_at) as date, COUNT(*) as count
		FROM url_clicks
		WHERE short_url_id = $1 
		AND clicked_at BETWEEN $2 AND $3
		GROUP BY DATE(clicked_at)
		ORDER BY date`
	
	getClicksByMonthQuery = `
		SELECT DATE_TRUNC('month', clicked_at) as month, COUNT(*) as count
		FROM url_clicks
		WHERE short_url_id = $1 
		AND clicked_at BETWEEN $2 AND $3
		GROUP BY DATE_TRUNC('month', clicked_at)
		ORDER BY month`
	
	getUserAgentStatsQuery = `
		SELECT user_agent, COUNT(*) as count
		FROM url_clicks
		WHERE short_url_id = $1 
		AND clicked_at BETWEEN $2 AND $3
		GROUP BY user_agent
		ORDER BY count DESC`
)

type AnalyticsRepository struct {
	PostgresDB *dbpg.DB
}

func NewAnalyticsRepository(config *config.Config) *AnalyticsRepository {
	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(config.MasterDSN, config.SlaveDSNs, opts)
	if err != nil {
		panic(err)
	}

	return &AnalyticsRepository{
		PostgresDB: db,
	}
}

func (a *AnalyticsRepository) RecordClick(ctx context.Context, click domain.URLClick) error {
	_, err := a.PostgresDB.ExecContext(ctx, recordClickQuery, click.ShortURLID, click.ClickedAt, click.IPAddress, click.UserAgent)
	return err
}

func (a *AnalyticsRepository) GetClickCount(ctx context.Context, shortUrlID string, startDate, endDate time.Time) (int, error) {
	var count int
	err := a.PostgresDB.QueryRowContext(ctx, getClickCountQuery, shortUrlID, startDate, endDate).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (a *AnalyticsRepository) GetClicksByPeriod(ctx context.Context, shortUrlID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error) {
	var query string
	if groupBy == "month" {
		query = getClicksByMonthQuery
	} else {
		query = getClicksByDayQuery
	}

	rows, err := a.PostgresDB.QueryContext(ctx, query, shortUrlID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}

	return results, nil
}

func (a *AnalyticsRepository) GetUserAgentStats(ctx context.Context, shortUrlID string, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	rows, err := a.PostgresDB.QueryContext(ctx, getUserAgentStatsQuery, shortUrlID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var userAgent string
		var count int
		if err := rows.Scan(&userAgent, &count); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"user_agent": userAgent,
			"count":      count,
		})
	}

	return results, nil
}
