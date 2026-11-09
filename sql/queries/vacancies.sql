-- name: ListVacancies :many
SELECT * FROM vacancies
ORDER BY order_index ASC;

-- name: ListActiveVacancies :many
SELECT * FROM vacancies
WHERE active = 1
ORDER BY order_index ASC;

-- name: CreateVacancy :one
INSERT INTO vacancies (
    id, title, department, salary, experience, schedule, responsibilities, requirements, benefits, active, order_index, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;
