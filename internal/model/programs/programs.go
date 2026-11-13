package programs

type Program struct {
	Executions []Execution
	Hour       Hour
}

type FertilizerProgram struct {
	Executions []Execution
	Hour       Hour
	Fertilizer bool
}

type Programs struct {
	Daily       []FertilizerProgram
	Odd         []FertilizerProgram
	Even        []FertilizerProgram
	Temperature Temperature
	Weekly      []Weekly
}
