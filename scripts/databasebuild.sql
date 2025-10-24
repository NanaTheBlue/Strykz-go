CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users(
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
username VARCHAR(322) UNIQUE NOT NULL,
email VARCHAR(255) UNIQUE NOT NULL,
hashed_password TEXT NOT NULL,
refresh_token TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

);



CREATE TABLE IF NOT EXISTS favor(
favor_id UUID PRIMARY KEY,
creator_id UUID REFERENCES users(id),
recipient_id UUID REFERENCES users(id),
favor_text TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);



CREATE TABLE IF NOT EXISTS relationship(
requester_id UUID REFERENCES users(id),
addressee_id UUID REFERENCES users(id),
relationship_status TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

PRIMARY KEY (requester_id, addressee_id),

CHECK (requester_id != addressee_id)



)


