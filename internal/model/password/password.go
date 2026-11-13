package password

import (
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	hash string
}

func (p *Password) Hash() string {
	return p.hash
}

func (p *Password) Compare(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(password))
}

func (p *Password) Hydrate(hash string) {
	p.hash = hash
}

func NewPassword(password string) (*Password, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Password{hash: string(bytes)}, nil
}
