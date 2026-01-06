-- Global Queries

-- name: GetLearnable :one
SELECT * FROM learnable
WHERE lid = ? LIMIT 1;

-- name: ListLearnables :many
SELECT * FROM learnable
ORDER BY lid;

-- name: CreateTraining :execresult
INSERT INTO training ( tstamp, speed, accy, nqtns, style ) VALUES (?, ?, ?, ?, ?);

-- name: CreateEncounter :execresult
INSERT INTO encounter ( qid, tstamp, estamp, entered, timer, correct, acty ) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetQuestionsInBand :many
SELECT l.lid, l.lname, l.lrank, l.defn, l.diffy, l.course,
       q.qtype, q.qid
FROM learnable l JOIN question q ON l.lid = q.lid
WHERE @low <= lrank AND lrank < @hi
ORDER BY random()
LIMIT @bandsize;


-- name: DeleteLearnable :exec
DELETE FROM learnable
WHERE lid = ?;
