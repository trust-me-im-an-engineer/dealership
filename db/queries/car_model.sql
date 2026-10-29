-- Операции со списком моделей автомобилей.
-- name: ListCarModels :many
SELECT m.name, m.year,
       (SELECT COUNT(*)
        FROM car_configuration c
        JOIN "order" o ON o.car_configuration_id = c.id
        WHERE c.car_model = m.name) AS order_count
FROM car_model m ORDER BY m.name;

-- name: GetCarModel :one
SELECT m.name, m.year,
       (SELECT COUNT(*)
        FROM car_configuration c
        JOIN "order" o ON o.car_configuration_id = c.id
        WHERE c.car_model = m.name) AS order_count
FROM car_model m WHERE m.name = $1;

-- name: CreateCarModel :exec
INSERT INTO car_model (name, year) VALUES ($1, $2);

-- name: UpdateCarModel :execrows
UPDATE car_model
SET name = sqlc.arg(new_name), year = sqlc.arg(year)
WHERE name = sqlc.arg(old_name);

-- name: DeleteCarModel :execrows
DELETE FROM car_model WHERE name = $1;

