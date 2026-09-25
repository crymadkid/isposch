package calendar

import (
	"strings"
	"testing"
	"time"

	"isposch/packages/schedule"
)

func TestGenerateEscapesFieldsAndProducesStableUID(t *testing.T) {
	lesson := schedule.Lesson{
		Date:     time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		Pair:     1,
		Subject:  "Математика, часть 1",
		Teacher:  "Иванов",
		Audience: "101",
		Note:     "Первая\rвторая",
		Campus:   "A;1",
	}

	first := Generate([]schedule.Lesson{lesson})
	second := Generate([]schedule.Lesson{lesson})

	if !strings.Contains(first, "SUMMARY:Математика\\, часть 1\r\n") {
		t.Fatalf("subject was not escaped in calendar:\n%s", first)
	}
	if !strings.Contains(first, "DESCRIPTION:Преподаватель: Иванов\\, Первая\\nвторая\r\n") {
		t.Fatalf("note was not escaped in calendar:\n%s", first)
	}
	if !strings.Contains(first, "LOCATION:101\\, корпус A\\;1\r\n") {
		t.Fatalf("location was not escaped in calendar:\n%s", first)
	}

	uid := func(value string) string {
		for _, line := range strings.Split(value, "\r\n") {
			if strings.HasPrefix(line, "UID:") {
				return line
			}
		}
		return ""
	}
	if uid(first) == "" || uid(first) != uid(second) {
		t.Fatalf("UID is not stable: %q vs %q", uid(first), uid(second))
	}
}
