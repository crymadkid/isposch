package parser

import (
	"encoding/xml"
	"io"
	"net/http"
)

// Фетчинг расписания с сайта
func FetchSchedule(url string) ([]Lesson, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseSchedule(body)
}

// Определяю структуры для парсинга XML
type dataroot struct {
	XMLName xml.Name `xml:"dataroot"`
	Lessons []Lesson `xml:"My"`
}

type Lesson struct {
	ID        int    `xml:"ID"`
	Date      string `xml:"DAT"`
	LessonNum int    `xml:"UR"`
	Teacher   string `xml:"FAMIO"`
	Subject   string `xml:"SPPRED.NAIM"`
	Group     string `xml:"SPGRUP.NAIM"`
	Room      string `xml:"AUD"`
	Campus    string `xml:"CAMPUS"`
}

func ParseSchedule(data []byte) ([]Lesson, error) {
	var result dataroot
	err := xml.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result.Lessons, nil
}
