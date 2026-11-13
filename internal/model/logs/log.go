package logs

import (
	"time"
)

type Log struct {
	ExecutedAt time.Time
	Seconds    int
	Zone       string
}
