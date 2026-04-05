-- name: CreateBooking :one
INSERT INTO bookings (
    id,
    slot_id,
    user_id,
    status,
    conference_link,
    created_at
) VALUES (
    @id,
    @slot_id,
    @user_id,
    @status,
    @conference_link,
    @created_at
)
RETURNING *;

-- name: GetBookingByID :one
SELECT
    id,
    slot_id,
    user_id,
    status,
    conference_link,
    created_at
FROM bookings
WHERE id = @id;

-- name: UpdateBookingStatus :exec
UPDATE bookings
SET status = @status
WHERE id = @id;

-- name: ListBookingsByUserID :many
SELECT
    b.id,
    b.slot_id,
    b.user_id,
    b.status,
    b.conference_link,
    b.created_at
FROM bookings b
JOIN slots s ON b.slot_id = s.id
WHERE b.user_id = @user_id
    AND s.start_time >= (NOW() AT TIME ZONE 'UTC')
ORDER BY s.start_time;

-- name: ListAllBookings :many
SELECT
    id,
    slot_id,
    user_id,
    status,
    conference_link,
    created_at,
    COUNT(*) OVER() AS total_count
FROM bookings
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
