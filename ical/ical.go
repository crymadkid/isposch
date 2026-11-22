package ical

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"isposch/parser"
)

var bells = map[int]string{
	1: "09:00",
	2: "10:45",
	3: "12:50",
	4: "14:30",
	5: "16:10",
}

const (
	lessonDuration = 90 * time.Minute
	utcLayout      = "20060102T150405Z"
	maxLineBytes   = 75
)

func Build(lessons []parser.Lesson, name string, loc *time.Location) (string, error) {
	var b strings.Builder
	write := func(line string) {
		b.WriteString(fold(line))
		b.WriteString("\r\n")
	}

	write("BEGIN:VCALENDAR")
	write("VERSION:2.0")
	write("PRODID:-//crymadkid//isposch//RU")
	write("CALSCALE:GREGORIAN")
	write("NAME:" + escape(name))
	write("X-WR-CALNAME:" + escape(name))

	stamp := time.Now().UTC().Format(utcLayout)

	for _, l := range lessons {
		start, err := lessonStart(l, loc)
		if err != nil {
			return "", fmt.Errorf("занятие %d: %w", l.ID, err)
		}
		end := start.Add(lessonDuration)

		write("BEGIN:VEVENT")
		write(fmt.Sprintf("UID:lesson-%d@isposch", l.ID))
		write("DTSTAMP:" + stamp)
		write("DTSTART:" + start.UTC().Format(utcLayout))
		write("DTEND:" + end.UTC().Format(utcLayout))
		write("SUMMARY:" + escape(l.Subject))
		write("DESCRIPTION:" + escape(description(l)))
		if place := location(l); place != "" {
			write("LOCATION:" + escape(place))
		}
		write("END:VEVENT")
	}

	write("END:VCALENDAR")
	return b.String(), nil
}

func lessonStart(l parser.Lesson, loc *time.Location) (time.Time, error) {
	day, err := l.Day(loc)
	if err != nil {
		return time.Time{}, err
	}

	hm, ok := bells[l.LessonNum]
	if !ok {
		return time.Time{}, fmt.Errorf("неизвестный номер пары: %d", l.LessonNum)
	}

	t, err := time.Parse("15:04", hm)
	if err != nil {
		return time.Time{}, fmt.Errorf("неверное время звонка %q: %w", hm, err)
	}

	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
}

// description формирует текст описания события.
func description(l parser.Lesson) string {
	parts := []string{
		"Преподаватель: " + l.Teacher,
		"Группа: " + l.Group,
	}
	if note := strings.TrimSpace(l.Note); note != "" {
		parts = append(parts, "Примечание: "+note)
	}
	return strings.Join(parts, "\n")
}

func location(l parser.Lesson) string {
	room := strings.TrimSpace(l.Room)
	campus := strings.TrimSpace(l.Campus)
	switch {
	case room != "" && campus != "":
		return room + " (" + campus + ")"
	case room != "":
		return room
	default:
		return campus
	}
}

func escape(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`;`, `\;`,
		`,`, `\,`,
		"\r\n", `\n`,
		"\n", `\n`,
	)
	return r.Replace(s)
}

func fold(line string) string {
	if len(line) <= maxLineBytes {
		return line
	}
	var b strings.Builder
	cur := 0
	for _, r := range line {
		size := utf8.RuneLen(r)
		if cur+size > maxLineBytes {
			b.WriteString("\r\n ")
			cur = 1
		}
		b.WriteRune(r)
		cur += size
	}
	return b.String()
}
