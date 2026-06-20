package port

import (
	"github.com/google/uuid"
)

type TokenGenerator interface {
	Generate(userID uuid.UUID, role string) (string, error)
	Validate(token string) (uuid.UUID, string, error)
}
