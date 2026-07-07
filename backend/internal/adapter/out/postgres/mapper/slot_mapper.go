package mapper

import (
	"github.com/maket12/meeting-rooms-api/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/meeting-rooms-api/internal/domain/model"

	"github.com/jackc/pgx/v5/pgtype"
)

func MapSlotsToSQLCCreateBatch(slots []*model.Slot) sqlc.CreateSlotsBatchParams {
	ids := make([]pgtype.UUID, len(slots))
	roomIDs := make([]pgtype.UUID, len(slots))
	startTimes := make([]pgtype.Timestamptz, len(slots))
	endTimes := make([]pgtype.Timestamptz, len(slots))

	for i, s := range slots {
		ids[i] = pgtype.UUID{Bytes: s.ID(), Valid: true}
		roomIDs[i] = pgtype.UUID{Bytes: s.RoomID(), Valid: true}
		startTimes[i] = pgtype.Timestamptz{Time: s.Start(), Valid: true}
		endTimes[i] = pgtype.Timestamptz{Time: s.End(), Valid: true}
	}

	return sqlc.CreateSlotsBatchParams{
		Ids:        ids,
		RoomIds:    roomIDs,
		StartTimes: startTimes,
		EndTimes:   endTimes,
	}
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
