-- +goose Up
-- +goose StatementBegin

-- Add CHECK constraint for version (IF NOT EXISTS not supported for constraints)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'databases_version_check'
    ) THEN
        ALTER TABLE databases
        ADD CONSTRAINT databases_version_check
        CHECK (version IN ('13', '14', '15', '16', '17', '18'));
    ELSE
        -- If constraint exists, drop and recreate with updated version list
        ALTER TABLE databases
        DROP CONSTRAINT databases_version_check,
        ADD CONSTRAINT databases_version_check
        CHECK (version IN ('13', '14', '15', '16', '17', '18'));
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the version CHECK constraint
ALTER TABLE databases
DROP CONSTRAINT IF EXISTS databases_version_check;

-- +goose StatementEnd
