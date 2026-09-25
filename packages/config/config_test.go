package config

import "testing"

func TestListenAddrPrefersValidPort(t *testing.T) {
	t.Setenv("PORT", "1234")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9999")

	if got := listenAddr(); got != ":1234" {
		t.Fatalf("listenAddr() = %q, want %q", got, ":1234")
	}
}

func TestListenAddrFallsBackForInvalidValues(t *testing.T) {
	t.Setenv("PORT", "invalid")
	t.Setenv("HTTP_ADDR", "also-invalid")

	if got := listenAddr(); got != ":8080" {
		t.Fatalf("listenAddr() = %q, want %q", got, ":8080")
	}
}
