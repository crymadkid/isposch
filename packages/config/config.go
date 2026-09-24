package config

import (
	"os"
	"strings"
)

const (
	defaultScheduleURL = "https://polytech-shedule.ru/data/2.xml"
	defaultGroup       = "25290901/3091"
)

type Config struct {
	ScheduleURL string
	Group       string
	HTTPAddr    string
}

func Load() Config {
	return Config{
		ScheduleURL: getenv("SCHEDULE_URL", defaultScheduleURL),
		Group:       getenv("SCHEDULE_GROUP", defaultGroup),
		HTTPAddr:    listenAddr(),
	}
}

func listenAddr() string {
	if port := getenv("PORT", ""); port != "" {
		// Render передает PORT числом, а Go ожидает формат ":порт".
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	return getenv("HTTP_ADDR", ":8080")
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
