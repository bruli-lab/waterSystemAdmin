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
	Daily       Daily
	Odd         Odd
	Even        Even
	Temperature Temperature
	Weekly      []Weekly
}
