-- name: GetLearnable :one
SELECT * FROM learnable
WHERE lid = ? LIMIT 1;

-- name: ListLearnables :many
SELECT * FROM learnable
ORDER BY lid;

-- name: CreateLearnable :execresult
INSERT INTO learnable (
  lid, esecs, timer, score, activity
) VALUES (
  ?, ?, ?, ?, ?
);

-- name: DeleteLearnable :exec
DELETE FROM learnable
WHERE lid = ?;
