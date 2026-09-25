package main

import (
	"log"
	"net/http"
	"time"

	"isposch/packages/config"
	"isposch/packages/server"
)

func main() {
	cfg := config.Load()
	client := &http.Client{Timeout: 30 * time.Second}

	handler := server.NewHandler(client, cfg.Group)

	log.Printf("calendar server listening on %s for group %s", cfg.HTTPAddr, cfg.Group)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		log.Fatal("ошибка запуска сервера: ", err)
	}
}

/*
	!!! оно к сожалению работает, живите с этим
*/
