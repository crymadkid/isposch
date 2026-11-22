package ical

import (
	"strings"
	"time"

	"isposch/parser"
)

func Build(lessons []parser.Lesson, name string, loc *time.Location) (string, error) {
	var b strings.Builder
	write := func(line string) {
		b.WriteString(line)
		b.WriteString("\r\n")
	}
	write("BEGIN:VCALENDAR")
	write("VERSION:2.0")
	write("PRODID:-//crymadkid//isposch//RU")
	write("END:VCALENDAR")
	return b.String(), nil
}
