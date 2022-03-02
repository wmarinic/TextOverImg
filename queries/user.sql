-- name: GetUser :one
SELECT *
FROM inspirationifier.users
WHERE user_name = $1;

-- name: CreateUser :one
INSERT INTO inspirationifier.users (
    User_Name, Password_Hash
) VALUES (
    $1, $2
) 
RETURNING *;