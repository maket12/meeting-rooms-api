-- name: CreateSlotsBatch :exec
INSERT INTO slots (
    id,
    room_id,
    start_time,
    end_time
)
SELECT
    unnest(@ids::uuid[]),
    unnest(@room_ids::uuid[]),
    unnest(@start_times::timestamptz[]),
    unnest(@end_times::timestamptz[])
ON CONFLICT (id) DO NOTHING;

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

-- name: HasSlotsForDate :one
SELECT EXISTS (
    SELECT 1
    FROM slots
    WHERE room_id = @room_id
        AND start_time::date = @date::date
);