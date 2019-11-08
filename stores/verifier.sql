CREATE TABLE IF NOT EXISTS VerificationRequests (
    autoID BIGSERIAL,
    id TEXT PRIMARY KEY,
    type TEXT,
    sender TEXT,
    recipient TEXT,
    data jsonb,
    secret TEXT NOT NULL,
    secretExpiry timestamptz NOT NULL,
    attempts integer,
    commStatus jsonb,
    status TEXT NOT NULL,
    createdAt timestamptz DEFAULT now(),
    updatedAt timestamptz DEFAULT now()
);
