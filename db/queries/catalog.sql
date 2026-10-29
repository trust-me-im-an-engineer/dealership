-- Операции со списком моделей автомобилей.
-- name: ListCarModels :many
SELECT name, year FROM car_model ORDER BY name;

-- name: GetCarModel :one
SELECT name, year FROM car_model WHERE name = $1;

-- name: CreateCarModel :exec
INSERT INTO car_model (name, year) VALUES ($1, $2);

-- name: UpdateCarModel :execrows
UPDATE car_model
SET name = sqlc.arg(new_name), year = sqlc.arg(year)
WHERE name = sqlc.arg(old_name);

-- name: DeleteCarModel :execrows
DELETE FROM car_model WHERE name = $1;

-- Операции со списком доступных комплектаций.
-- name: ListCarConfigurations :many
SELECT id, name, car_model, price::TEXT AS price, delivery_days, description
FROM car_configuration ORDER BY car_model, name, id;

-- name: GetCarConfiguration :one
SELECT id, name, car_model, price::TEXT AS price, delivery_days, description
FROM car_configuration WHERE id = $1;

-- name: CreateCarConfiguration :exec
INSERT INTO car_configuration (name, car_model, price, delivery_days, description)
VALUES (sqlc.arg(name), sqlc.arg(car_model), sqlc.arg(price)::NUMERIC,
        sqlc.arg(delivery_days), sqlc.narg(description));

-- name: UpdateCarConfiguration :execrows
UPDATE car_configuration
SET name = sqlc.arg(name), car_model = sqlc.arg(car_model), price = sqlc.arg(price)::NUMERIC,
    delivery_days = sqlc.arg(delivery_days), description = sqlc.narg(description)
WHERE id = sqlc.arg(id);

-- name: DeleteCarConfiguration :execrows
DELETE FROM car_configuration WHERE id = $1;
