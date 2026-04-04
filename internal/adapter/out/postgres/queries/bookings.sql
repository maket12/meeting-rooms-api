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
    id,
    slot_id,
    user_id,
    status,
    conference_link,
    created_at
FROM bookings
WHERE user_id = @user_id;

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
