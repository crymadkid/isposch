package parser

import "strings"

// Filter возвращает занятия, для которых keep вернул true.
func Filter(lessons []Lesson, keep func(Lesson) bool) []Lesson {
	filtered := make([]Lesson, 0)
	for _, l := range lessons {
		if keep(l) {
			filtered = append(filtered, l)
		}
	}
	return filtered
}

// FilterByGroup возвращает занятия группы (без учёта регистра и пробелов по краям).
func FilterByGroup(lessons []Lesson, group string) []Lesson {
	group = strings.TrimSpace(group)
	return Filter(lessons, func(l Lesson) bool {
		return strings.EqualFold(strings.TrimSpace(l.Group), group)
	})
}
