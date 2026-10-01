-- Create resolved_media table
CREATE TABLE resolved_media (
    id UUID PRIMARY KEY,
    source_url TEXT NOT NULL,
    platform TEXT NOT NULL,
    media_type TEXT NOT NULL,
    title TEXT,
    thumbnail_url TEXT,
    width INTEGER,
    height INTEGER,
    duration INTEGER,
    formats JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- RLS
ALTER TABLE resolved_media ENABLE ROW LEVEL SECURITY;
-- Keep access entirely restricted to service_role (no public policies needed)
