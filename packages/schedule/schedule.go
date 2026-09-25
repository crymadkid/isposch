package schedule

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type scheduleXML struct {
	Generated string      `xml:"generated,attr"`
	Lessons   []lessonXML `xml:"My"`
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

const dateLayout = "2006-01-02T15:04:05"

const (
	scheduleBaseURL       = "https://polytech-shedule.ru/data/"
	scheduleFileCount     = 20
	scheduleCacheDuration = 15 * time.Minute
)

var scheduleCache struct {
	sync.Mutex
	url        string
	discovered time.Time
}

func FetchLessons(ctx context.Context, client *http.Client, group string) ([]Lesson, error) {
	if client == nil {
		return nil, errors.New("клиент расписания не задан")
	}
	group = strings.TrimSpace(group)
	if group == "" {
		return nil, errors.New("группа не задана")
	}

	resolvedURL, err := resolveScheduleURL(ctx, client)
	if err != nil {
		return nil, err
	}

	document, err := fetchDocument(ctx, client, resolvedURL)
	if err != nil {
		var statusErr *httpStatusError
		if errors.As(err, &statusErr) && statusErr.Code == http.StatusNotFound {
			invalidateScheduleURL(resolvedURL)
			resolvedURL, err = resolveScheduleURL(ctx, client)
			if err == nil {
				document, err = fetchDocument(ctx, client, resolvedURL)
			}
		}
	}
	if err != nil {
		return nil, err
	}

	lessons := make([]Lesson, 0)
	for _, item := range document.Lessons {
		if strings.TrimSpace(item.Group) != group {
			continue
		}

		date, err := time.Parse(dateLayout, strings.TrimSpace(item.Date))
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

func resolveScheduleURL(ctx context.Context, client *http.Client) (string, error) {
	scheduleCache.Lock()
	defer scheduleCache.Unlock()
	if scheduleCache.url != "" && time.Since(scheduleCache.discovered) < scheduleCacheDuration {
		return scheduleCache.url, nil
	}

	resolvedURL, err := discoverScheduleURL(ctx, client)
	if err != nil {
		return "", err
	}
	scheduleCache.url = resolvedURL
	scheduleCache.discovered = time.Now()
	return resolvedURL, nil
}

func invalidateScheduleURL(url string) {
	scheduleCache.Lock()
	defer scheduleCache.Unlock()
	if scheduleCache.url == url {
		scheduleCache.url = ""
		scheduleCache.discovered = time.Time{}
	}
}

func discoverScheduleURL(ctx context.Context, client *http.Client) (string, error) {
	var newestURL string
	var newestDate time.Time

	for fileNumber := 1; fileNumber <= scheduleFileCount; fileNumber++ {
		url := scheduleBaseURL + strconv.Itoa(fileNumber) + ".xml"
		document, err := fetchDocument(ctx, client, url)
		if err != nil {
			var statusErr *httpStatusError
			if errors.As(err, &statusErr) && statusErr.Code == http.StatusNotFound {
				continue
			}
			return "", err
		}

		generated, err := time.Parse(dateLayout, strings.TrimSpace(document.Generated))
		if err != nil {
			return "", fmt.Errorf("не удалось разобрать дату обновления %q в %s: %w", document.Generated, url, err)
		}
		if newestURL == "" || generated.After(newestDate) {
			newestURL = url
			newestDate = generated
		}
	}

	if newestURL == "" {
		return "", fmt.Errorf("не найдено ни одного XML-файла расписания")
	}
	return newestURL, nil
}

func fetchDocument(ctx context.Context, client *http.Client, url string) (scheduleXML, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return scheduleXML{}, fmt.Errorf("не удалось создать запрос расписания: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return scheduleXML{}, fmt.Errorf("не удалось скачать расписание: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return scheduleXML{}, &httpStatusError{Code: resp.StatusCode, Status: resp.Status}
	}

	var document scheduleXML
	if err := xml.NewDecoder(io.LimitReader(resp.Body, 32<<20)).Decode(&document); err != nil {
		return scheduleXML{}, fmt.Errorf("не удалось прочитать XML расписания: %w", err)
	}
	return document, nil
}

type httpStatusError struct {
	Code   int
	Status string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("сервер расписания вернул HTTP %s", e.Status)
}
