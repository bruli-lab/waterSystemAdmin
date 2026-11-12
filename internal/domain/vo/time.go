package vo

import (
	"fmt"
	"strconv"
	"time"
	_ "time/tzdata"
)

var madrid = loadMadrid()

func loadMadrid() *time.Location {
	loc, err := time.LoadLocation("Europe/Madrid")
	if err != nil {
		return time.UTC
	}
	return loc
}

func ParseTimeFromUnix(s string) (time.Time, error) {
	segs, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
	}
	return time.Unix(segs, 0).In(madrid), nil
}
