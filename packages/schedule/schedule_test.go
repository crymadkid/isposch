package schedule

import (
	"context"
	"net/http"
	"testing"
)

func TestFetchLessonsRejectsInvalidInput(t *testing.T) {
	if _, err := FetchLessons(context.Background(), nil, "group"); err == nil {
		t.Fatal("FetchLessons() accepted a nil client")
	}
	if _, err := FetchLessons(context.Background(), &http.Client{}, " "); err == nil {
		t.Fatal("FetchLessons() accepted an empty group")
	}
}
