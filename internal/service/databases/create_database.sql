-- name: DatabasesServiceCreateDatabase :one
INSERT INTO databases (
  name, connection_string, database_type, version
)
VALUES (
  @name, pgp_sym_encrypt(@connection_string, @encryption_key), @database_type, @version
)
RETURNING *;
