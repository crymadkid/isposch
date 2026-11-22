// Package parser загружает и разбирает XML-расписание.
package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	dateLayout   = "2006-01-02T15:04:05"
	maxBodyBytes = 10 << 20 // 10 МБ
)

var client = &http.Client{
	Timeout: 15 * time.Second,
}

type dataroot struct {
	XMLName xml.Name `xml:"dataroot"`
	Lessons []Lesson `xml:"My"`
}

// Lesson описывает одно занятие из расписания.
type Lesson struct {
	ID        int    `xml:"ID"`
	Date      string `xml:"DAT"`
	LessonNum int    `xml:"UR"`
	Teacher   string `xml:"FAMIO"`
	Subject   string `xml:"SPPRED.NAIM"`
	Group     string `xml:"SPGRUP.NAIM"`
	Room      string `xml:"AUD"`
	Note      string `xml:"NOTE"`
	Campus    string `xml:"CAMPUS"`
	Zam       int    `xml:"ZAM"`
	Kons      int    `xml:"KONS"`
}

// Day возвращает дату занятия в указанном часовом поясе.
func (l Lesson) Day(loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(dateLayout, l.Date, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("неверная дата %q: %w", l.Date, err)
	}
	return t, nil
}

// FetchSchedule загружает XML по URL и разбирает его.
func FetchSchedule(url string) ([]Lesson, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("запрос расписания: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP ошибка: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("чтение ответа: %w", err)
	}
	return ParseSchedule(body)
}

// ParseSchedule разбирает XML-расписание.
func ParseSchedule(data []byte) ([]Lesson, error) {
	var result dataroot
	if err := xml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("разбор XML: %w", err)
	}
	return result.Lessons, nil
}
