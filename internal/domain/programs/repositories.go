package programs

import "context"

type Repository interface {
	FindAll(ctx context.Context) (*Programs, error)
	Save(ctx context.Context, p *FertilizerProgram, t *TypeProgram) error
	Remove(ctx context.Context, hour *Hour, t *TypeProgram) error
}

type WeeklyRepository interface {
	Save(ctx context.Context, p *Weekly) error
	Remove(ctx context.Context, day *WeekDay) error
}

type TemperatureRepository interface {
	Save(ctx context.Context, p *TemperatureProgram) error
	Remove(ctx context.Context, temperature int) error
}
