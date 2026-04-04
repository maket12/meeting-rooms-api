-- name: CreateSchedule :one
INSERT INTO schedules (
    id,
    room_id,
    days_of_week,
    start_minutes,
    end_minutes
) VALUES (
    @id,
    @room_id,
    @days_of_week,
    @start_minutes,
    @end_minutes
)
RETURNING *;

-- name: GetScheduleByRoomID :one
SELECT
    id,
    room_id,
    days_of_week,
    start_minutes,
    end_minutes
FROM schedules
WHERE room_id = @room_id;