package ical

import (
	"strings"
	"testing"
	"time"

	"isposch/parser"
)

// useTestBells подменяет таблицу звонков на время теста.
func useTestBells(t *testing.T) {
	t.Helper()
	old := bells
	bells = map[int]string{3: "12:50"}
	t.Cleanup(func() { bells = old })
}

func TestBuild_Empty(t *testing.T) {
	got, err := Build([]parser.Lesson{}, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if !strings.HasPrefix(got, "BEGIN:VCALENDAR\r\n") {
		t.Errorf("должен начинаться с BEGIN:VCALENDAR, получено %q", got)
	}
	if !strings.HasSuffix(got, "END:VCALENDAR\r\n") {
		t.Errorf("должен заканчиваться END:VCALENDAR, получено %q", got)
	}
	if strings.Contains(got, "BEGIN:VEVENT") {
		t.Error("в пустом календаре не должно быть событий")
	}
}

func TestBuild_Header(t *testing.T) {
	got, err := Build(nil, "Расписание 1234", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	for _, line := range []string{
		"VERSION:2.0",
		"PRODID:-//crymadkid//isposch//RU",
		"NAME:Расписание 1234",
		"X-WR-CALNAME:Расписание 1234",
	} {
		if !strings.Contains(got, line+"\r\n") {
			t.Errorf("в заголовке нет строки %q\nполучено:\n%s", line, got)
		}
	}
}

func TestBuild_OneLesson(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{
		ID:        1,
		Date:      "2026-09-10T00:00:00",
		LessonNum: 3,
		Teacher:   "Иванов И.И.",
		Subject:   "Предмет",
		Group:     "1234",
		Room:      "101",
		Campus:    "Э",
	}}

	got, err := Build(lessons, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	got = strings.ReplaceAll(got, "\r\n ", "") // <-- сюда

	wantLines := []string{
		"BEGIN:VEVENT",
		"UID:lesson-1@isposch",
		"DTSTART:20260910T125000Z",
		"DTEND:20260910T142000Z", // +90 минут
		"SUMMARY:Предмет",
		`DESCRIPTION:Преподаватель: Иванов И.И.\nГруппа: 1234`,
		"LOCATION:101 (Э)",
		"END:VEVENT",
	}
	for _, line := range wantLines {
		if !strings.Contains(got, line+"\r\n") {
			t.Errorf("нет строки %q\nполучено:\n%s", line, got)
		}
	}
}

func TestBuild_Note(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{
		ID: 1, Date: "2026-09-10T00:00:00", LessonNum: 3,
		Teacher: "Иванов И.И.", Group: "1234", Note: "Зачёт",
	}}
	got, err := Build(lessons, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	got = strings.ReplaceAll(got, "\r\n ", "") // <-- сюда

	if !strings.Contains(got, `\nПримечание: Зачёт`) {
		t.Errorf("примечание не попало в описание:\n%s", got)
	}
}

func TestBuild_MultipleLessons(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{
		{ID: 1, Date: "2026-09-10T00:00:00", LessonNum: 3, Subject: "A"},
		{ID: 2, Date: "2026-09-11T00:00:00", LessonNum: 3, Subject: "B"},
	}
	got, err := Build(lessons, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if n := strings.Count(got, "BEGIN:VEVENT"); n != 2 {
		t.Errorf("ожидалось 2 события, получено %d", n)
	}
	if !strings.Contains(got, "UID:lesson-1@isposch") || !strings.Contains(got, "UID:lesson-2@isposch") {
		t.Errorf("UID разных занятий должны различаться:\n%s", got)
	}
}

func TestBuild_Timezone(t *testing.T) {
	useTestBells(t)

	msk := time.FixedZone("MSK", 3*60*60)
	lessons := []parser.Lesson{{ID: 1, Date: "2026-09-10T00:00:00", LessonNum: 3, Subject: "X"}}

	got, err := Build(lessons, "Тест", msk)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	// 12:50 по Москве = 09:50 UTC
	if !strings.Contains(got, "DTSTART:20260910T095000Z\r\n") {
		t.Errorf("время не переведено в UTC:\n%s", got)
	}
}

func TestBuild_NoLocationWhenEmpty(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{ID: 1, Date: "2026-09-10T00:00:00", LessonNum: 3, Subject: "X"}}
	got, err := Build(lessons, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if strings.Contains(got, "LOCATION:") {
		t.Errorf("LOCATION не должен выводиться без аудитории и корпуса:\n%s", got)
	}
}

func TestBuild_EscapesSubject(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{ID: 1, Date: "2026-09-10T00:00:00", LessonNum: 3, Subject: "Физика, лекция; 1"}}
	got, err := Build(lessons, "Тест", time.UTC)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	if !strings.Contains(got, `SUMMARY:Физика\, лекция\; 1`+"\r\n") {
		t.Errorf("спецсимволы не экранированы:\n%s", got)
	}
}

func TestBuild_UnknownBell(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{ID: 7, Date: "2026-09-10T00:00:00", LessonNum: 99}}
	if _, err := Build(lessons, "Тест", time.UTC); err == nil {
		t.Fatal("ожидалась ошибка на неизвестном номере пары")
	}
}

func TestBuild_BadDate(t *testing.T) {
	useTestBells(t)

	lessons := []parser.Lesson{{ID: 7, Date: "мусор", LessonNum: 3}}
	if _, err := Build(lessons, "Тест", time.UTC); err == nil {
		t.Fatal("ожидалась ошибка на неверной дате")
	}
}

func TestLessonStart(t *testing.T) {
	useTestBells(t)

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

func TestLessonStart_UnknownBell(t *testing.T) {
	useTestBells(t)

	l := parser.Lesson{Date: "2026-09-10T00:00:00", LessonNum: 99}
	if _, err := lessonStart(l, time.UTC); err == nil {
		t.Fatal("ожидалась ошибка на неизвестном номере пары")
	}
}

func TestLocation(t *testing.T) {
	tests := []struct {
		name string
		l    parser.Lesson
		want string
	}{
		{"аудитория и корпус", parser.Lesson{Room: "101", Campus: "Э"}, "101 (Э)"},
		{"только аудитория", parser.Lesson{Room: "101"}, "101"},
		{"только корпус", parser.Lesson{Campus: "Э"}, "Э"},
		{"ничего", parser.Lesson{}, ""},
		{"пробелы", parser.Lesson{Room: " 101 ", Campus: " Э "}, "101 (Э)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := location(tt.l); got != tt.want {
				t.Errorf("получено %q, ожидалось %q", got, tt.want)
			}
		})
	}
}

func TestEscape(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"без спецсимволов", "abc", "abc"},
		{"запятая и точка с запятой", "a,b;c", `a\,b\;c`},
		{"обратный слэш", `a\b`, `a\\b`},
		{"перевод строки", "a\nb", `a\nb`},
		{"CRLF", "a\r\nb", `a\nb`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escape(tt.in); got != tt.want {
				t.Errorf("получено %q, ожидалось %q", got, tt.want)
			}
		})
	}
}

func TestFold(t *testing.T) {
	t.Run("короткая строка не меняется", func(t *testing.T) {
		if got := fold("SUMMARY:abc"); got != "SUMMARY:abc" {
			t.Errorf("получено %q", got)
		}
	})

	t.Run("длинная строка с кириллицей", func(t *testing.T) {
		in := "SUMMARY:" + strings.Repeat("Привет", 30)
		got := fold(in)

		for _, line := range strings.Split(got, "\r\n") {
			if len(line) > maxLineBytes {
				t.Errorf("строка длиннее %d байт: %d", maxLineBytes, len(line))
			}
		}
		if unfolded := strings.ReplaceAll(got, "\r\n ", ""); unfolded != in {
			t.Errorf("после склейки получено %q, ожидалось %q", unfolded, in)
		}
	})
}
