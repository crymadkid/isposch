package schedule

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

func FetchLessons(ctx context.Context, client *http.Client, url, group string) ([]Lesson, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать запрос расписания: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("не удалось скачать расписание: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервер расписания вернул HTTP %s", resp.Status)
	}

	var document scheduleXML
	if err := xml.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&document); err != nil {
		return nil, fmt.Errorf("не удалось прочитать XML расписания: %w", err)
	}

	lessons := make([]Lesson, 0)
	for _, item := range document.Lessons {
		if strings.TrimSpace(item.Group) != group {
			continue
		}

		date, err := time.Parse("2006-01-02T15:04:05", strings.TrimSpace(item.Date))
		if err != nil {
			return nil, fmt.Errorf("не удалось разобрать дату занятия %q: %w", item.Date, err)
		}
		if _, ok := PairSlots(date, item.Pair); !ok {
			return nil, fmt.Errorf("неподдерживаемый номер пары: %d", item.Pair)
		}

		lessons = append(lessons, Lesson{
			Date: date, Pair: item.Pair, Teacher: strings.TrimSpace(item.Teacher),
			Subject: strings.TrimSpace(item.Subject), Audience: strings.TrimSpace(item.Audience),
			Note: strings.TrimSpace(item.Note), Campus: strings.TrimSpace(item.Campus),
		})
	}

	return lessons, nil
}
