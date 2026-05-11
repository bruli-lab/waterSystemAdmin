package execution

import "context"

type Clean struct {
	repo Repository
}

func (e *Clean) Execute(ctx context.Context, id string) error {
	return e.repo.SendClean(ctx, id)
}

func NewClean(repo Repository) *Clean {
	return &Clean{repo: repo}
}
