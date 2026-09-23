package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultScheduleURL = "https://polytech-shedule.ru/data/2.xml"
	defaultGroup       = "25290901/3091"
)

type scheduleXML struct {
	Lessons []lessonXML `xml:"My"`
}

type lessonXML struct {
	Date     string `xml:"DAT"`
	Pair     int    `xml:"UR"`
	Teacher  string `xml:"FAMIO"`
	Subject  string `xml:"SPPRED.NAIM"`
	Group    string `xml:"SPGRUP.NAIM"`
	Audience string `xml:"AUD"`
	Note     string `xml:"NOTE"`
	Campus   string `xml:"CAMPUS"`
}

type Lesson struct {
	Date     time.Time
	Pair     int
	Teacher  string
	Subject  string
	Audience string
	Note     string
	Campus   string
}

type pairTime struct {
	start, end string
}

var weekdayPairSlots = map[int][]pairTime{
	1: {{"09:00", "10:30"}},
	2: {{"10:45", "11:30"}, {"11:55", "12:40"}},
	3: {{"12:50", "14:20"}},
	4: {{"14:30", "16:00"}},
	5: {{"16:10", "17:40"}},
}

var weekdayPairTimes = map[int]pairTime{
	1: {"09:00", "10:30"},
	2: {"10:45", "12:40"},
	3: {"12:50", "14:20"},
	4: {"14:30", "16:00"},
	5: {"16:10", "17:40"},
}

var saturdayPairTimes = map[int]pairTime{
	1: {"09:00", "10:30"},
	2: {"10:40", "12:10"},
	3: {"12:30", "14:00"},
	4: {"14:10", "15:40"},
}

func fetchLessons(ctx context.Context, client *http.Client, url, group string) ([]Lesson, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create schedule request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download schedule: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("schedule returned HTTP %s", resp.Status)
	}

	var document scheduleXML
	if err := xml.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&document); err != nil {
		return nil, fmt.Errorf("decode schedule XML: %w", err)
	}

	lessons := make([]Lesson, 0)
	for _, item := range document.Lessons {
		if strings.TrimSpace(item.Group) != group {
			continue
		}
		date, err := time.Parse("2006-01-02T15:04:05", strings.TrimSpace(item.Date))
		if err != nil {
			return nil, fmt.Errorf("parse lesson date %q: %w", item.Date, err)
		}
		if _, ok := pairSlots(date, item.Pair); !ok {
			return nil, fmt.Errorf("unsupported pair number %d", item.Pair)
		}
		lessons = append(lessons, Lesson{
			Date: date, Pair: item.Pair, Teacher: strings.TrimSpace(item.Teacher),
			Subject: strings.TrimSpace(item.Subject), Audience: strings.TrimSpace(item.Audience),
			Note: strings.TrimSpace(item.Note), Campus: strings.TrimSpace(item.Campus),
		})
	}
	return lessons, nil
}

func calendar(lessons []Lesson) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//isposch//Polytech schedule//RU\r\nCALSCALE:GREGORIAN\r\nMETHOD:PUBLISH\r\nX-WR-CALNAME:Расписание 25290901/3091\r\nX-WR-TIMEZONE:Europe/Moscow\r\n")
	for _, lesson := range lessons {
		slots, _ := pairSlots(lesson.Date, lesson.Pair)
		for slotIndex, slot := range slots {
			start := atTime(lesson.Date, slot.start)
			end := atTime(lesson.Date, slot.end)
			uidInput := fmt.Sprintf("%s|%d|%d|%s|%s", lesson.Date.Format("2006-01-02"), lesson.Pair, slotIndex, lesson.Subject, lesson.Audience)
			sum := sha256.Sum256([]byte(uidInput))
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

func atTime(date time.Time, value string) time.Time {
	hour, _ := strconv.Atoi(value[:2])
	minute, _ := strconv.Atoi(value[3:])
	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, 0, 0, time.FixedZone("Europe/Moscow", 3*60*60))
}

func pairTimes(date time.Time, pair int) (pairTime, bool) {
	slots, ok := pairSlots(date, pair)
	if !ok {
		return pairTime{}, false
	}
	return slots[0], true
}

func pairSlots(date time.Time, pair int) ([]pairTime, bool) {
	if date.Weekday() == time.Saturday {
		slot, ok := saturdayPairTimes[pair]
		if !ok {
			return nil, false
		}
		return []pairTime{slot}, true
	}
	slots, ok := weekdayPairSlots[pair]
	return slots, ok
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
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, ";", `\;`)
	value = strings.ReplaceAll(value, ",", `\,`)
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", `\n`), "\n", `\n`)
}

func main() {
	scheduleURL := getenv("SCHEDULE_URL", defaultScheduleURL)
	group := getenv("SCHEDULE_GROUP", defaultGroup)
	client := &http.Client{Timeout: 30 * time.Second}

	http.HandleFunc("/calendar.ics", func(w http.ResponseWriter, r *http.Request) {
		lessons, err := fetchLessons(r.Context(), client, scheduleURL, group)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=900")
		_, _ = io.WriteString(w, calendar(lessons))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	addr := listenAddr()
	log.Printf("calendar server listening on %s for group %s", addr, group)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func listenAddr() string {
	if port := getenv("PORT", ""); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	return getenv("HTTP_ADDR", ":8080")
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
