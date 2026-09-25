package config

import (
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	defaultGroup = "25290901/3091"
)

type Config struct {
	Group    string
	HTTPAddr string
}

func Load() Config {
	return Config{
		Group:    getenv("SCHEDULE_GROUP", defaultGroup),
		HTTPAddr: listenAddr(),
	}
}

func listenAddr() string {
	if port := getenv("PORT", ""); port != "" {
		// Render передает PORT числом, а Go ожидает формат ":порт".
		port = strings.TrimPrefix(port, ":")
		if number, err := strconv.Atoi(port); err == nil && number >= 1 && number <= 65535 {
			return ":" + strconv.Itoa(number)
		}
	}
	addr := getenv("HTTP_ADDR", ":8080")
	if _, _, err := net.SplitHostPort(addr); err == nil {
		return addr
	}
	if strings.HasPrefix(addr, ":") {
		if number, err := strconv.Atoi(strings.TrimPrefix(addr, ":")); err == nil && number >= 1 && number <= 65535 {
			return addr
		}
	}
	return ":8080"
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
