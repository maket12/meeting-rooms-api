package port

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	EnsureDummyUsers(ctx context.Context, adminID, userID uuid.UUID) error
}
