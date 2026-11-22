package ical

import (
	"strings"
	"testing"
	"time"

	"isposch/parser"
)

func TestBuild(t *testing.T) {
	got, err := Build([]parser.Lesson{}, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if !strings.Contains(got, "BEGIN:VCALENDAR") {
		t.Errorf("ожидался BEGIN:VCALENDAR, но не найден")
	}
	if !strings.Contains(got, "END:VCALENDAR") {
		t.Errorf("ожидался END:VCALENDAR, но не найден")
	}
}

func TestLessonStart(t *testing.T) {
	l := parser.Lesson{Date: "2026-09-10T00:00:00", LessonNum: 3}

	got, err := lessonStart(l, time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	want := time.Date(2026, 9, 10, 12, 50, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("получено %v, ожидалось %v", got, want)
	}
}
