-- name: GetUserByUsername :one
SELECT * FROM `users` WHERE username = ? LIMIT 1;


alter table project add column