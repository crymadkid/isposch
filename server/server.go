package server

import (
	"isposch/ical"
	"isposch/parser"
	"log"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	Source string
	Loc    *time.Location
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/schedule.ics", s.ServeSchedule)
	return mux
}

func (s *Server) ServeSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	group := strings.TrimSpace(r.URL.Query().Get("group"))
	if group == "" {
		http.Error(w, "Не указана группа", http.StatusBadRequest)
		return
	}

	lessons, err := parser.FetchSchedule(s.Source)
	if err != nil {
		log.Printf("Загрузка расписания: %v", err)
		http.Error(w, "Ошибка загрузки расписания", http.StatusBadGateway)
		return
	}
	lessons = parser.FilterByGroup(lessons, group)

	cal, err := ical.Build(lessons, "Расписание "+group, s.Loc)
	if err != nil {
		log.Printf("Сборка календаря: %v", err)
		http.Error(w, "Ошибка сборки календаря", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Write([]byte(cal))
}
