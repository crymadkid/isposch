package parser

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func readTestdata(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/schedule.xml")
	if err != nil {
		t.Fatalf("не удалось прочитать testdata: %v", err)
	}
	return data
}

func TestParseSchedule(t *testing.T) {
	lessons, err := ParseSchedule(readTestdata(t))
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if len(lessons) != 1 {
		t.Fatalf("ожидалась 1 запись, получено %d", len(lessons))
	}

	want := Lesson{
		ID:        1,
		Date:      "2026-09-10T00:00:00",
		LessonNum: 1,
		Teacher:   "Иванов И.И.",
		Subject:   "Предмет",
		Group:     "1234",
		Campus:    "Э",
		Zam:       2,
	}
	if lessons[0] != want {
		t.Errorf("получено %+v, ожидалось %+v", lessons[0], want)
	}
}

func TestParseSchedule_InvalidXML(t *testing.T) {
	if _, err := ParseSchedule([]byte(`<dataroot"></dataroot>`)); err == nil {
		t.Fatal("ожидалась ошибка на битом XML")
	}
}

func TestParseSchedule_Empty(t *testing.T) {
	lessons, err := ParseSchedule([]byte(`<dataroot></dataroot>`))
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if len(lessons) != 0 {
		t.Errorf("ожидалось 0 записей, получено %d", len(lessons))
	}
}

func TestLessonDay(t *testing.T) {
	got, err := Lesson{Date: "2026-09-10T00:00:00"}.Day(time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("получено %v, ожидалось %v", got, want)
	}

	if _, err := (Lesson{Date: "мусор"}).Day(time.UTC); err == nil {
		t.Error("ожидалась ошибка на неверной дате")
	}
}

func TestFetchSchedule(t *testing.T) {
	data := readTestdata(t)

	t.Run("успех", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(data)
		}))
		defer srv.Close()

		lessons, err := FetchSchedule(srv.URL)
		if err != nil {
			t.Fatalf("ошибка: %v", err)
		}
		if len(lessons) != 1 {
			t.Fatalf("ожидалась 1 запись, получено %d", len(lessons))
		}
	})

	t.Run("статус 500", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}))
		defer srv.Close()

		if _, err := FetchSchedule(srv.URL); err == nil {
			t.Fatal("ожидалась ошибка при статусе 500")
		}
	})
}

func TestFilterByGroup(t *testing.T) {
	lessons := []Lesson{
		{ID: 1, Group: "A"},
		{ID: 2, Group: "B"},
		{ID: 3, Group: " a "},
	}

	tests := []struct {
		name  string
		group string
		want  int
	}{
		{"точное совпадение", "A", 2},
		{"другая группа", "B", 1},
		{"регистр и пробелы", " a", 2},
		{"нет совпадений", "C", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByGroup(lessons, tt.group)
			if len(got) != tt.want {
				t.Errorf("получено %d, ожидалось %d", len(got), tt.want)
			}
		})
	}
}
