CREATE TABLE IF NOT EXISTS tasks(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    content text NOT NULL,
    is_completed boolean NOT NULL,
    version integer NOT NULL DEFAULT 1 
);