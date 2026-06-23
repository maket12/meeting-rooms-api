package port

import (
	"backend/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *model.Schedule) (*model.Schedule, error)
	Get(ctx context.Context, roomID uuid.UUID) (*model.Schedule, error)
}
