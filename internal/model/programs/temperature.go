package programs

type (
	Temperature        []TemperatureProgram
	TemperatureProgram struct {
		Programs    []Program
		Temperature int
	}
)
