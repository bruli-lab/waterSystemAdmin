package zones

import (
	"context"
)

type FindZones struct {
	repo Repository
}

func (z FindZones) Find(ctx context.Context) ([]Zone, error) {
	return z.repo.FindAll(ctx)
}

func NewFindZones(repo Repository) *FindZones {
	return &FindZones{repo: repo}
}
