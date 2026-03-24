CREATE INDEX IF NOT EXISTS lists_title_index ON lists USING GIN (to_tsvector('simple', title));
CREATE INDEX IF NOT EXISTS tasks_content_index ON tasks USING GIN (to_tsvector('simple', content));