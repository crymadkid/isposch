package server

import (
	"io"
	"net/http"

	"isposch/packages/calendar"
	"isposch/packages/schedule"
)

func NewHandler(client *http.Client, scheduleURL, group string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/calendar.ics", func(w http.ResponseWriter, r *http.Request) {
		lessons, err := schedule.FetchLessons(r.Context(), client, scheduleURL, group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=900")
		_, _ = io.WriteString(w, calendar.Generate(lessons))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
