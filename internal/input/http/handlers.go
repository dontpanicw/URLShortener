package http

import (
	"context"
	"encoding/json"
	"github.com/dontpanicw/URLShortener/internal/port"
	"log"
	"net/http"
	"strings"
	"time"
)

type Handler struct {
	urlUsecases       port.UrlUsecases
	analyticsUsecases port.AnalyticsUsecases
}

func NewHandler(urlUsecases port.UrlUsecases, analyticsUsecases port.AnalyticsUsecases) *Handler {
	return &Handler{
		urlUsecases:       urlUsecases,
		analyticsUsecases: analyticsUsecases,
	}
}

type ShortenRequest struct {
	Url         string `json:"url"`
	CustomShort string `json:"custom_short,omitempty"`
}

type ShortenResponse struct {
	ShortUrl string `json:"short_url"`
	FullUrl  string `json:"full_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Вспомогательные функции
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// POST /shorten - создание короткой ссылки
func (h *Handler) ShortenUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Url == "" {
		writeError(w, http.StatusBadRequest, "URL is required")
		return
	}

	url, err := h.urlUsecases.CreateShortUrl(r.Context(), req.Url, req.CustomShort)
	if err != nil {
		log.Printf("Error creating short URL: %v", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ShortenResponse{
		ShortUrl: url.ShortUrl,
		FullUrl:  url.FullUrl,
	})
}

// GET /{short_url} - редирект на оригинальный URL
func (h *Handler) RedirectUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	shortUrl := strings.TrimPrefix(r.URL.Path, "/")
	if shortUrl == "" {
		writeError(w, http.StatusBadRequest, "Short URL is required")
		return
	}

	fullUrl, err := h.urlUsecases.GetOriginalUrl(r.Context(), shortUrl)
	if err != nil {
		log.Printf("Error getting original URL: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if fullUrl == "" {
		writeError(w, http.StatusNotFound, "Short URL not found")
		return
	}

	// Записываем клик асинхронно
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.urlUsecases.RecordClick(ctx, shortUrl, ipAddress, userAgent); err != nil {
			log.Printf("Error recording click: %v", err)
		}
	}()

	http.Redirect(w, r, fullUrl, http.StatusMovedPermanently)
}

// GET /analytics/{short_url} - получение аналитики
func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем short_url из пути /analytics/{short_url}
	path := strings.TrimPrefix(r.URL.Path, "/analytics/")
	parts := strings.Split(path, "/")
	shortUrl := parts[0]

	if shortUrl == "" {
		writeError(w, http.StatusBadRequest, "Short URL is required")
		return
	}

	// Параметры для фильтрации по датам
	query := r.URL.Query()
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
	} else {
		endDate = time.Now()
	}

	stats, err := h.analyticsUsecases.GetClickStats(r.Context(), shortUrl, startDate, endDate)
	if err != nil {
		log.Printf("Error getting analytics: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GET /analytics/{short_url}/user-agents - статистика по User-Agent
func (h *Handler) GetUserAgentStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Извлекаем short_url из пути /analytics/{short_url}/user-agents
	path := strings.TrimPrefix(r.URL.Path, "/analytics/")
	path = strings.TrimSuffix(path, "/user-agents")
	shortUrl := path

	if shortUrl == "" {
		writeError(w, http.StatusBadRequest, "Short URL is required")
		return
	}

	query := r.URL.Query()
	startDateStr := query.Get("start_date")
	endDateStr := query.Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid start_date format")
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid end_date format")
			return
		}
	} else {
		endDate = time.Now()
	}

	stats, err := h.analyticsUsecases.GetUserAgentStats(r.Context(), shortUrl, startDate, endDate)
	if err != nil {
		log.Printf("Error getting user agent stats: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"short_url":        shortUrl,
		"user_agent_stats": stats,
	})
}

// GET /popular - получение популярных ссылок
func (h *Handler) GetPopularUrls(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	limit := 10 // По умолчанию топ-10
	urls, err := h.urlUsecases.GetPopularUrls(r.Context(), limit)
	if err != nil {
		log.Printf("Error getting popular URLs: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"popular_urls": urls,
		"count":        len(urls),
	})
}

// GET /cache/stats - статистика кэша
func (h *Handler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.urlUsecases.GetCacheStats(r.Context())
	if err != nil {
		log.Printf("Error getting cache stats: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
