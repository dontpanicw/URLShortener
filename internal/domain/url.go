package domain

import "time"

type Url struct {
	Id        string    `json:"id"`
	FullUrl   string    `json:"full_url"`
	ShortUrl  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

// Сущность "Клик/переход"
type URLClick struct {
	Id         string    `json:"id"`
	ShortURLID string    `json:"short_url_id"` // внешний ключ
	ClickedAt  time.Time `json:"clicked_at"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
}
