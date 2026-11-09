-- name: ListChatMessagesBySession :many
SELECT * FROM chat_messages
WHERE session_id = ?
ORDER BY created_at ASC;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (
    id, session_id, sender, message, tg_msg_id, created_at
) VALUES (
    ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
RETURNING *;
