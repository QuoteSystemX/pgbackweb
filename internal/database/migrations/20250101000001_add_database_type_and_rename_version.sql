-- +goose Up
-- +goose StatementBegin

-- Add database_type column with default 'postgresql' for existing rows
ALTER TABLE databases
ADD COLUMN IF NOT EXISTS database_type TEXT NOT NULL DEFAULT 'postgresql';

-- Add CHECK constraint for database_type (IF NOT EXISTS not supported for constraints)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'databases_database_type_check'
    ) THEN
        ALTER TABLE databases
        ADD CONSTRAINT databases_database_type_check
        CHECK (database_type IN ('postgresql', 'clickhouse'));
    END IF;
END $$;

-- Rename pg_version to version
ALTER TABLE databases
RENAME COLUMN pg_version TO version;

-- Drop the old pg_version CHECK constraint
ALTER TABLE databases
DROP CONSTRAINT IF EXISTS databases_pg_version_check;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rename version back to pg_version
ALTER TABLE databases
RENAME COLUMN version TO pg_version;

-- Restore the old pg_version CHECK constraint
ALTER TABLE databases
ADD CONSTRAINT databases_pg_version_check
CHECK (pg_version IN ('13', '14', '15', '16', '17', '18'));

-- Drop the database_type CHECK constraint
ALTER TABLE databases
DROP CONSTRAINT IF EXISTS databases_database_type_check;

-- Remove database_type column
ALTER TABLE databases
DROP COLUMN IF EXISTS database_type;

-- +goose StatementEnd
