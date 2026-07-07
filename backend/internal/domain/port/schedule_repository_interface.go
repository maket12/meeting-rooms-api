package port

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error)
	Get(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}
