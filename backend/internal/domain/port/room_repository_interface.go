package port

import (
	"context"

	"github.com/maket12/meeting-rooms-api/internal/domain/model"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) (*model.Room, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Room, error)
	List(ctx context.Context) ([]*model.Room, error)
}
