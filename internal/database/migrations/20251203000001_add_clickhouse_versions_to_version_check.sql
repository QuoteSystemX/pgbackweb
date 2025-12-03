-- +goose Up
-- +goose StatementBegin

-- Update CHECK constraint to allow both PostgreSQL and ClickHouse versions
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
        version IN (
            -- PostgreSQL versions
            '13', '14', '15', '16', '17', '18',
            -- ClickHouse versions
            '22.8', '23.8', '24.1', '24.3'
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

-- +goose StatementEnd

