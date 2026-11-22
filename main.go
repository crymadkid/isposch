package main

import (
	"flag"
	"isposch/server"
	"log"
	"net/http"
	"time"
	_ "time/tzdata"
)

func main() {
	addr := flag.String("addr", ":8080", "адрес сервера")
	source := flag.String("source", "", "URL расписания")
	tz := flag.String("tz", "Europe/Moscow", "часовой пояс")
	flag.Parse()

	if *source == "" {
		log.Fatal("не указан URL расписания")
	}
	loc, err := time.LoadLocation(*tz)
	if err != nil {
		log.Fatalf("не удалось загрузить часовой пояс %q: %v", *tz, err)
	}

	s := &server.Server{Source: *source, Loc: loc}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      s.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("Слушаю %s", *addr)
	log.Fatal(srv.ListenAndServe())
}
