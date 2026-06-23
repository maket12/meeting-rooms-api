-- name: CreateSlotsBatch :copyfrom
INSERT INTO slots (
    id,
    room_id,
    start_time,
    end_time
) VALUES (
    @id,
    @room_id,
    @start_time,
    @end_time
);

-- name: GetSlotByID :one
SELECT
    id,
    room_id,
    start_time,
    end_time
FROM slots
WHERE id = @id;

-- name: GetFreeSlotsByRoomAndDate :many
SELECT
    s.id,
    s.room_id,
    s.start_time,
    s.end_time
FROM slots s
LEFT JOIN bookings b ON s.id = b.slot_id AND b.status = 'active'
WHERE s.room_id = @room_id
    AND s.start_time::date = @date::date
    AND b.id IS NULL
ORDER BY s.start_time;