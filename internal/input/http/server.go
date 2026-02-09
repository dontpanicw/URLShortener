package http

import (
	"github.com/dontpanicw/URLShortener/internal/port"
	"net/http"
	"strings"
)

func NewServer(urlUsecases port.UrlUsecases, analyticsUsecases port.AnalyticsUsecases) http.Handler {
	handler := NewHandler(urlUsecases, analyticsUsecases)
	mux := http.NewServeMux()

	// Статические файлы для UI
	fs := http.FileServer(http.Dir("./web"))
	
	// API routes
	mux.HandleFunc("/shorten", handler.ShortenUrl)
	mux.HandleFunc("/analytics/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/user-agents") {
			handler.GetUserAgentStats(w, r)
		} else {
			handler.GetAnalytics(w, r)
		}
	})

	// Главный обработчик
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Если путь начинается с /analytics или /shorten, пропускаем
		if strings.HasPrefix(r.URL.Path, "/analytics") || r.URL.Path == "/shorten" {
			http.NotFound(w, r)
			return
		}

		// Если это корень или файл из web/, отдаем статику
		if r.URL.Path == "/" || strings.Contains(r.URL.Path, ".") {
			fs.ServeHTTP(w, r)
			return
		}

		// Иначе это короткая ссылка - делаем редирект
		handler.RedirectUrl(w, r)
	})

	return mux
}
