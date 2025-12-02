-- +goose Up
-- +goose StatementBegin

-- Make version column nullable
ALTER TABLE databases
ALTER COLUMN version DROP NOT NULL;

-- Update CHECK constraint to allow NULL for ClickHouse
-- For PostgreSQL, version must be one of the supported versions
-- For ClickHouse, version can be NULL
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'databases_version_check'
    ) THEN
        ALTER TABLE databases
        DROP CONSTRAINT databases_version_check;
    END IF;
    
    ALTER TABLE databases
    ADD CONSTRAINT databases_version_check
    CHECK (
        version IS NULL 
        OR (
            database_type = 'postgresql' 
            AND version IN ('13', '14', '15', '16', '17', '18')
        )
    );
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the old CHECK constraint (only PostgreSQL versions)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'databases_version_check'
    ) THEN
        ALTER TABLE databases
        DROP CONSTRAINT databases_version_check;
    END IF;
    
    ALTER TABLE databases
    ADD CONSTRAINT databases_version_check
    CHECK (version IN ('13', '14', '15', '16', '17', '18'));
END $$;

-- Make version NOT NULL again
ALTER TABLE databases
ALTER COLUMN version SET NOT NULL;

-- +goose StatementEnd

