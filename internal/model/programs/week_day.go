package programs

import (
	"errors"
	"time"
)

var ErrInvalidWeekDay = errors.New("invalid week day")

type WeekDay time.Weekday

func (d WeekDay) String() string {
	return time.Weekday(d).String()
}

var days = map[string]time.Weekday{
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
	"Sunday":    time.Sunday,
}

func ParseWeekDay(s string) (WeekDay, error) {
	day, ok := days[s]
	if !ok {
		return 0, ErrInvalidWeekDay
	}
	return WeekDay(day), nil
}
