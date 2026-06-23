package port

import (
	"backend/internal/domain/model"
	"context"
	"time"

	"github.com/google/uuid"
)

type SlotRepository interface {
	CreateBatch(ctx context.Context, slots []*model.Slot) error
	Get(ctx context.Context, id uuid.UUID) (*model.Slot, error)
	ListFree(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*model.Slot, error)
}
