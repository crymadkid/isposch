package calendar

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"isposch/packages/schedule"
)

func Generate(lessons []schedule.Lesson) string {
	var b strings.Builder

	// * Начало календаря
	b.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//isposch//Polytech schedule//RU\r\nCALSCALE:GREGORIAN\r\nMETHOD:PUBLISH\r\nX-WR-CALNAME:Расписание ИСПО\r\nX-WR-TIMEZONE:Europe/Moscow\r\n")

	for _, lesson := range lessons {
		slots, _ := schedule.PairSlots(lesson.Date, lesson.Pair)
		for slotIndex, slot := range slots {
			start := schedule.AtTime(lesson.Date, slot.Start)
			end := schedule.AtTime(lesson.Date, slot.End)

			uidInput := fmt.Sprintf(
				"%s|%d|%d|%s|%s",
				lesson.Date.Format("2006-01-02"),
				lesson.Pair,
				slotIndex,
				lesson.Subject,
				lesson.Audience,
			)
			sum := sha256.Sum256([]byte(uidInput))

			// * Аудитория и корпус идут в одном поле, чтобы календарь нормально их показывал
			location := lesson.Audience
			if lesson.Campus != "" {
				location = joinParts(location, "корпус "+lesson.Campus)
			}
			description := joinParts("Преподаватель: "+lesson.Teacher, lesson.Note)

			b.WriteString("BEGIN:VEVENT\r\n")
			fmt.Fprintf(&b, "UID:%s@isposch\r\n", hex.EncodeToString(sum[:]))
			fmt.Fprintf(&b, "DTSTAMP:%s\r\n", time.Now().UTC().Format("20060102T150405Z"))
			fmt.Fprintf(&b, "DTSTART;TZID=Europe/Moscow:%s\r\n", start.Format("20060102T150405"))
			fmt.Fprintf(&b, "DTEND;TZID=Europe/Moscow:%s\r\n", end.Format("20060102T150405"))
			fmt.Fprintf(&b, "SUMMARY:%s\r\n", icsEscape(lesson.Subject))
			fmt.Fprintf(&b, "LOCATION:%s\r\n", icsEscape(location))
			fmt.Fprintf(&b, "DESCRIPTION:%s\r\n", icsEscape(description))
			b.WriteString("END:VEVENT\r\n")
		}
	}

	b.WriteString("END:VCALENDAR\r\n")
	return b.String()
}

func joinParts(parts ...string) string {
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			values = append(values, strings.TrimSpace(part))
		}
	}
	return strings.Join(values, ", ")
}

func icsEscape(value string) string {
	// * В iCalendar некоторые символы имеют специальное значение, поэтому их нужно экранировать спасибо гуглу за подсказку
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, ";", `\;`)
	value = strings.ReplaceAll(value, ",", `\,`)
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", `\n`), "\n", `\n`)
}
