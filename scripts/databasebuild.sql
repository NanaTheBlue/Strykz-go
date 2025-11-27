CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users(
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
username VARCHAR(322) UNIQUE NOT NULL,
email VARCHAR(255) UNIQUE NOT NULL,
hashed_password TEXT NOT NULL,
refresh_token TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

);





CREATE TABLE IF NOT EXISTS notifications(
requester_id UUID REFERENCES users(id) ON DELETE CASCADE,
addressee_id UUID REFERENCES users(id) ON DELETE CASCADE,
notification_status TEXT NOT NULL,
notification_type TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

PRIMARY KEY (requester_id, addressee_id),

CHECK (requester_id != addressee_id)



);


CREATE TABLE IF NOT EXISTS friends (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id <> friend_id),
    CHECK (user_id < friend_id);
);


CREATE TABLE IF NOT EXISTS blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (blocker_id, blocked_id),
    CHECK (blocker_id <> blocked_id)
);





