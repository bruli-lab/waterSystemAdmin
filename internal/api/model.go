package api

type Log struct {
	ExecutedAt string `json:"executed_at"`
	Seconds    int    `json:"seconds"`
	ZoneName   string `json:"zone_name"`
}

type Status struct {
	SystemStartedAt string `json:"system_started_at"`
	Temperature     float64
	Humidity        float64
	IsRaining       bool   `json:"is_raining"`
	IsDay           bool   `json:"is_day"`
	UpdatedAt       string `json:"updated_at"`
	Active          bool
	Clean           bool
}

type Program struct {
	Executions []Execution `json:"executions"`
	Hour       string      `json:"hour"`
}
type FertilizerProgram struct {
	Executions []Execution `json:"executions"`
	Hour       string      `json:"hour"`
	Fertilizer bool        `json:"fertilizer"`
}

type ProgramWeekly struct {
	WeekDay  string    `json:"week_day"`
	Programs []Program `json:"programs"`
}

type Programs struct {
	Daily       []FertilizerProgram  `json:"daily"`
	Temperature []TemperatureProgram `json:"temperature"`
	Odd         []FertilizerProgram  `json:"odd"`
	Even        []FertilizerProgram  `json:"even"`
	Weekly      []ProgramWeekly      `json:"weekly"`
}

type Execution struct {
	Seconds int      `json:"seconds"`
	Zones   []string `json:"zones"`
}

type TemperatureProgram struct {
	Programs    []Program `json:"programs"`
	Temperature int       `json:"temperature"`
}

type UpdateFertilizerProgram struct {
	Executions []Execution `json:"executions"`
	Fertilizer bool        `json:"fertilizer"`
}

type UpdatePrograms struct {
	Programs []Program `json:"programs"`
}

type ExecuteZone struct {
	Seconds int `json:"seconds"`
}

type Zone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Relays []int  `json:"relays"`
}
type UpdateZone struct {
	Name   string `json:"name"`
	Relays []int  `json:"relays"`
}
