package migrations

import (
	"database/sql"
	"log"
)

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS urls (
			id VARCHAR(255) PRIMARY KEY,
			full_url TEXT NOT NULL,
			short_url VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_short_url ON urls(short_url)`,
		`CREATE TABLE IF NOT EXISTS url_clicks (
			id SERIAL PRIMARY KEY,
			short_url_id VARCHAR(255) NOT NULL REFERENCES urls(id) ON DELETE CASCADE,
			clicked_at TIMESTAMP NOT NULL DEFAULT NOW(),
			ip_address VARCHAR(45),
			user_agent TEXT,
			CONSTRAINT fk_short_url FOREIGN KEY (short_url_id) REFERENCES urls(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_clicks_short_url ON url_clicks(short_url_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clicks_date ON url_clicks(clicked_at)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("Migration error: %v", err)
			return err
		}
	}

	log.Println("All migrations applied successfully")
	return nil
}
