-- name: CreateRoom :one
INSERT INTO rooms (
    id,
    name,
    description,
    capacity,
    created_at
) VALUES (
    @id, @name, @description, @capacity, @created_at
)
RETURNING *;

-- name: GetRoomByID :one
SELECT
    id,
    name,
    description,
    capacity,
    created_at
FROM rooms
WHERE id = @id;

-- name: ListRooms :many
SELECT
    id,
    name,
    description,
    capacity,
    created_at
FROM rooms;