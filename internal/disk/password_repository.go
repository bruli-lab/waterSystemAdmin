package disk

import (
	"context"
	"errors"
	"io/fs"
	"os"

	"github.com/bruli/waterSystemAdmin/internal/model/password"
)

type PasswordRepository struct {
	filePath string
}

func (p PasswordRepository) Read(ctx context.Context) (*password.Password, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		hash, err := os.ReadFile(p.filePath)
		if err != nil {
			return nil, err
		}
		var pass password.Password
		pass.Hydrate(string(hash))
		return &pass, nil
	}
}

func (p PasswordRepository) Save(ctx context.Context, pass *password.Password) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return os.WriteFile(p.filePath, []byte(pass.Hash()), 0o600)
	}
}

func (p PasswordRepository) Exists(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		_, err := os.Stat(p.filePath)
		switch {
		case err == nil:
			return true, nil
		case errors.Is(err, fs.ErrNotExist):
			return false, nil
		default:
			return false, err
		}
	}
}

func NewPasswordRepository(filePath string) *PasswordRepository {
	return &PasswordRepository{filePath: filePath}
}
