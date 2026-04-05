package port

import (
	"MeetingRoomsAPI/internal/domain/model"
	"context"

	"github.com/google/uuid"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *model.Booking) (*model.Booking, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Booking, error)
	ListAll(ctx context.Context, limit, offset int32) ([]*model.Booking, int64, error)
}
