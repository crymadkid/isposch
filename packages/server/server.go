package server

import (
	"io"
	"net/http"

	"isposch/packages/calendar"
	"isposch/packages/schedule"
)

func NewHandler(client *http.Client, group string) http.Handler {
	mux := http.NewServeMux()

	calendarHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		lessons, err := schedule.FetchLessons(r.Context(), client, group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=900")
		_, _ = io.WriteString(w, calendar.Generate(lessons))
	}

	// Любой путь возвращает один и тот же календарь: это удобно для URL-подписок.
	mux.HandleFunc("/", calendarHandler)
	return mux
}
