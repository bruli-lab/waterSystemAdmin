package execution

import "context"

type Repository interface {
	SendExecution(ctx context.Context, exe *Execution) error
	SendFertilizer(ctx context.Context, exe *Execution) error
	SendClean(ctx context.Context, id string) error
}
