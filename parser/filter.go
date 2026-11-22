package parser

import "strings"

func Filter(lessons []Lesson, keep func(Lesson) bool) []Lesson {
	filtered := make([]Lesson, 0)
	for _, l := range lessons {
		if keep(l) {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

func FilterByGroup(lessons []Lesson, group string) []Lesson {
	group = strings.TrimSpace(group)
	return Filter(lessons, func(l Lesson) bool {
		return strings.EqualFold(strings.TrimSpace(l.Group), group)
	})
}
