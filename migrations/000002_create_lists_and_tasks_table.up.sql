CREATE TABLE IF NOT EXISTS lists(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title text NOT NULL,
    description text NOT NULL,
    version integer NOT NULL DEFAULT 1 
);

CREATE TABLE IF NOT EXISTS tasks(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    list_id UUID REFERENCES lists(id) ON DELETE CASCADE,
    content text NOT NULL,
    is_completed boolean NOT NULL,
    version integer NOT NULL DEFAULT 1 
);