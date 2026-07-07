package port

import (
	"context"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"
	"time"

	"github.com/google/uuid"
)

type SlotRepository interface {
	CreateBatch(ctx context.Context, slots []*model.Slot) error
	Get(ctx context.Context, id uuid.UUID) (*model.Slot, error)
	ListFree(ctx context.Context, roomID uuid.UUID, date time.Time) ([]*model.Slot, error)
	ExistsForDate(ctx context.Context, roomID uuid.UUID, date time.Time) (bool, error)
}
