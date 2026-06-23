package port

import (
	"backend/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) (*model.Room, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Room, error)
	List(ctx context.Context) ([]*model.Room, error)
}
