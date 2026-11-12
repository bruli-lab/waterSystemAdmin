package programs

import "errors"

const (
	DailyProgramType TypeProgram = "daily"
	OddProgramType   TypeProgram = "odd"
	EvenProgramType  TypeProgram = "even"
)

var programTypeMap = map[string]TypeProgram{
	"daily": DailyProgramType,
	"odd":   OddProgramType,
	"even":  EvenProgramType,
}

type TypeProgram string

func (t TypeProgram) String() string {
	return string(t)
}

func ParseProgramType(s string) (*TypeProgram, error) {
	p, ok := programTypeMap[s]
	if !ok {
		return nil, errors.New("invalid program type")
	}
	return &p, nil
}
