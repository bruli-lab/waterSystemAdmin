package programs

import "context"

type Create struct {
	repo Repository
}

func (c Create) Create(ctx context.Context, p *FertilizerProgram, t *TypeProgram) error {
	return c.repo.Save(ctx, p, t)
}

func NewCreate(repo Repository) *Create {
	return &Create{repo: repo}
}
