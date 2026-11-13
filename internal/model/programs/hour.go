package programs

import (
	"errors"
	"time"
)

var ErrInvalidExecutionHour = errors.New("invalid execution hour")

type Hour time.Time

func (h Hour) String() string {
	return time.Time(h).Format("15:04")
}

func ParseHour(s string) (Hour, error) {
	h, err := time.Parse("15:04", s)
	if err != nil {
		return Hour{}, ErrInvalidExecutionHour
	}
	return Hour(h), nil
}
