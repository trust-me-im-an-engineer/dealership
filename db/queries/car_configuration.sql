-- Операции со списком доступных комплектаций.
-- name: ListCarConfigurations :many
SELECT c.id, c.name, c.car_model, c.price::TEXT AS price, c.delivery_days, c.description,
       (SELECT COUNT(*) FROM "order" o WHERE o.car_configuration_id = c.id) AS order_count
FROM car_configuration c ORDER BY c.car_model, c.name, c.id;

-- name: GetCarConfiguration :one
SELECT c.id, c.name, c.car_model, c.price::TEXT AS price, c.delivery_days, c.description,
       (SELECT COUNT(*) FROM "order" o WHERE o.car_configuration_id = c.id) AS order_count
FROM car_configuration c WHERE c.id = $1;

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
