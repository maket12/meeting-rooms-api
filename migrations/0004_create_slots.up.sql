--- Table of slots
CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,

    CONSTRAINT unique_room_slot_time UNIQUE (room_id, start_time)
);

CREATE INDEX IF NOT EXISTS idx_slots_filter ON slots(room_id, start_time);