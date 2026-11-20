CREATE INDEX IF NOT EXISTS ix_task_started_at_desc ON task(started_at DESC NULLS LAST);
