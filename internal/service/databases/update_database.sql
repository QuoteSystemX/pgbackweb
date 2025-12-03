-- name: DatabasesServiceUpdateDatabase :one
UPDATE databases
SET
  name = COALESCE(sqlc.narg('name'), name),
  database_type = COALESCE(sqlc.narg('database_type'), database_type),
  version = COALESCE(sqlc.narg('version'), version),
  connection_string = CASE
    WHEN sqlc.narg('connection_string')::TEXT IS NOT NULL
    THEN pgp_sym_encrypt(
      sqlc.narg('connection_string')::TEXT, sqlc.arg('encryption_key')::TEXT
    )
    ELSE connection_string
  END
WHERE id = @id
RETURNING *;
