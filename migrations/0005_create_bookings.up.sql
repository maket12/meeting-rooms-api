-- Table of bookings
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id UUID NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK ( status in ('active', 'cancelled') ),
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_slot
ON bookings (slot_id)
WHERE (status = 'active');

CREATE INDEX IF NOT EXISTS idx_bookings_user_active ON bookings (user_id) WHERE (status = 'active');