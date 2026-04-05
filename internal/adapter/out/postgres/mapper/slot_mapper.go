package mapper

import (
	"MeetingRoomsAPI/internal/adapter/out/postgres/sqlc"
	"MeetingRoomsAPI/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapSlotsToSQLCCreateBatch(slots []*model.Slot) []sqlc.CreateSlotsBatchParams {
	mapped := make([]sqlc.CreateSlotsBatchParams, len(slots))
	for i := range mapped {
		mapped[i] = sqlc.CreateSlotsBatchParams{
			ID: pgtype.UUID{
				Bytes: slots[i].ID(),
				Valid: true,
			},
			RoomID: pgtype.UUID{
				Bytes: slots[i].RoomID(),
				Valid: true,
			},
			StartTime: pgtype.Timestamptz{
				Time:             slots[i].Start(),
				InfinityModifier: 0,
				Valid:            true,
			},
			EndTime: pgtype.Timestamptz{
				Time:             slots[i].End(),
				InfinityModifier: 0,
				Valid:            true,
			},
		}
	}
	return mapped
}

func MapSQLCToSlot(rawSlot sqlc.Slot) *model.Slot {
	return model.RestoreSlot(
		rawSlot.ID.Bytes,
		rawSlot.RoomID.Bytes,
		rawSlot.StartTime.Time.UTC(),
		rawSlot.EndTime.Time.UTC(),
	)
}

func MapSQLCToSlots(rawSlots []sqlc.Slot) []*model.Slot {
	slots := make([]*model.Slot, len(rawSlots))
	for i := range slots {
		slots[i] = MapSQLCToSlot(rawSlots[i])
	}
	return slots
}
