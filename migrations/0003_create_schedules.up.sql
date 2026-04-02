-- Table of schedules
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    days_of_week INTEGER[] NOT NULL,
    start_minutes INTEGER NOT NULL,  -- Minutes from the start of the day
    end_minutes INTEGER NOT NULL,    -- Minutes from the end of the day

    -- To ensure there is only one schedule for each room
    CONSTRAINT unique_room_schedule UNIQUE (room_id)
);

CREATE INDEX idx_schedules_room_id ON schedules(room_id);