package ical

import (
	"fmt"
	"strings"
	"time"

	"isposch/parser"
)

var bells = map[int]string{
	1: "09:00",
	2: "10:45",
	3: "12:50",
	4: "14:30",
	5: "16:10",
}

const lessonDuration = 90 * time.Minute

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
		return time.Time{}, err
	}

	return time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
}

func Build(lessons []parser.Lesson, name string, loc *time.Location) (string, error) {
	var b strings.Builder
	write := func(line string) {
		b.WriteString(line)
		b.WriteString("\r\n")
	}
	write("BEGIN:VCALENDAR")
	write("VERSION:2.0")
	write("PRODID:-//crymadkid//isposch//RU")
	write("NAME:" + name)
	write("END:VCALENDAR")
	return b.String(), nil
}
