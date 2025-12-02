-- name: DatabasesServiceUpdateDatabase :one
UPDATE databases
SET
  name = COALESCE(sqlc.narg('name'), name),
  database_type = COALESCE(sqlc.narg('database_type'), database_type),
  version = CASE
    WHEN sqlc.narg('version')::TEXT IS NULL THEN version
    WHEN sqlc.narg('version')::TEXT = '' THEN NULL
    ELSE sqlc.narg('version')::TEXT
  END,
  connection_string = CASE
    WHEN sqlc.narg('connection_string')::TEXT IS NOT NULL
    THEN pgp_sym_encrypt(
      sqlc.narg('connection_string')::TEXT, sqlc.arg('encryption_key')::TEXT
    )
    ELSE connection_string
  END
WHERE id = @id
RETURNING *;
