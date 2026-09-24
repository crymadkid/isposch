package schedule

import (
	"strconv"
	"time"
)

type PairTime struct {
	Start string
	End   string
}

var weekdayPairSlots = map[int][]PairTime{
	1: {{"09:00", "10:30"}},
	2: {{"10:45", "12:40"}},
	3: {{"12:50", "14:20"}},
	4: {{"14:30", "16:00"}},
	5: {{"16:10", "17:40"}},
}

var saturdayPairTimes = map[int]PairTime{
	1: {"09:00", "10:30"},
	2: {"10:40", "12:10"},
	3: {"12:30", "14:00"},
	4: {"14:10", "15:40"},
}

func PairSlots(date time.Time, pair int) ([]PairTime, bool) {
	if date.Weekday() == time.Saturday {
		// В субботу используется отдельное расписание.
		slot, ok := saturdayPairTimes[pair]
		if !ok {
			return nil, false
		}
		return []PairTime{slot}, true
	}

	slots, ok := weekdayPairSlots[pair]
	return slots, ok
}

func AtTime(date time.Time, value string) time.Time {
	hour, _ := strconv.Atoi(value[:2])
	minute, _ := strconv.Atoi(value[3:])
	return time.Date(
		date.Year(), date.Month(), date.Day(), hour, minute, 0, 0,
		time.FixedZone("Europe/Moscow", 3*60*60),
	)
}
