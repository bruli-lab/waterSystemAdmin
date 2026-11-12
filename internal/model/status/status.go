package status

import "time"

type Status struct {
	SystemStartedAt time.Time
	Temperature     float64
	Humidity        float64
	IsRaining       bool
	IsDay           bool
	UpdatedAt       time.Time
	Active          bool
	Clean           bool
}
