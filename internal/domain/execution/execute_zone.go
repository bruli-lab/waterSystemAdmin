package execution

import "context"

type ExecuteZone struct {
	repo Repository
}

func (e ExecuteZone) Execute(ctx context.Context, exe *Execution) error {
	if exe.Fertilizer {
		return e.repo.SendFertilizer(ctx, exe)
	}
	return e.repo.SendExecution(ctx, exe)
}

func NewExecuteZone(repo Repository) *ExecuteZone {
	return &ExecuteZone{repo: repo}
}
