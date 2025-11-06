-- name: ExecutionsServiceGetExecution :one
SELECT
  executions.*,
  databases.id AS database_id,
  databases.database_type AS database_database_type,
  databases.version AS database_version
FROM executions
INNER JOIN backups ON backups.id = executions.backup_id
INNER JOIN databases ON databases.id = backups.database_id
WHERE executions.id = @id;
