package port

import (
	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, role string) (string, error)
	ValidateToken(token string) (uuid.UUID, string, error)
}
