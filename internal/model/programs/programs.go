package programs

type Program struct {
	Executions []Execution
	Hour       Hour
	Enabled    bool
}

type FertilizerProgram struct {
	Executions []Execution
	Hour       Hour
	Fertilizer bool
	Enabled    bool
}

type Programs struct {
	Daily       []FertilizerProgram
	Odd         []FertilizerProgram
	Even        []FertilizerProgram
	Temperature Temperature
	Weekly      []Weekly
}
