package parser

func FilterByGroup(lessons []Lesson, group string) []Lesson {
	var filtered []Lesson
	for _, lesson := range lessons {
		if lesson.Group == group {
			filtered = append(filtered, lesson)
		}
	}
	return filtered
}
