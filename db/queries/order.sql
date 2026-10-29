-- Заказы для просмотра администратором.
-- name: ListOrders :many
SELECT o.id,
       TO_CHAR(o.ordered_at, 'DD.MM.YYYY HH24:MI') AS ordered_at,
       c.car_model,
       c.name AS configuration_name,
       o.buyer_name,
       o.buyer_phone,
       o.buyer_email,
       u.username AS seller,
       TO_CHAR(o.expected_delivery_on, 'DD.MM.YYYY') AS expected_delivery_on,
       CASE o.status
           WHEN 'placed' THEN 'Оформлен'
           WHEN 'delivered' THEN 'Доставлен'
           WHEN 'cancelled' THEN 'Отменён'
       END::TEXT AS status
FROM "order" o
JOIN car_configuration c ON c.id = o.car_configuration_id
JOIN "user" u ON u.id = o.seller_id
ORDER BY o.ordered_at DESC, o.id DESC;

-- Заказы текущего продавца для главной страницы.
-- name: ListSellerOrders :many
SELECT o.id,
       TO_CHAR(o.ordered_at, 'DD.MM.YYYY HH24:MI') AS ordered_at,
       c.car_model,
       c.name AS configuration_name,
       o.buyer_name,
       o.buyer_phone,
       o.buyer_email,
       TO_CHAR(o.expected_delivery_on, 'DD.MM.YYYY') AS expected_delivery_on,
       CASE o.status
           WHEN 'placed' THEN 'Оформлен'
           WHEN 'delivered' THEN 'Доставлен'
           WHEN 'cancelled' THEN 'Отменён'
       END::TEXT AS status
FROM "order" o
JOIN car_configuration c ON c.id = o.car_configuration_id
WHERE o.seller_id = $1
ORDER BY o.ordered_at DESC, o.id DESC;

-- Заказ текущего продавца для заполнения строки редактирования.
-- name: GetSellerOrder :one
SELECT o.id,
       TO_CHAR(o.ordered_at, 'DD.MM.YYYY HH24:MI') AS ordered_at,
       c.car_model,
       c.name AS configuration_name,
       o.car_configuration_id,
       o.buyer_name,
       o.buyer_phone,
       o.buyer_email,
       TO_CHAR(o.expected_delivery_on, 'DD.MM.YYYY') AS expected_delivery_on,
       TO_CHAR(o.expected_delivery_on, 'YYYY-MM-DD') AS expected_delivery_on_input,
       CASE o.status
           WHEN 'placed' THEN 'Оформлен'
           WHEN 'delivered' THEN 'Доставлен'
           WHEN 'cancelled' THEN 'Отменён'
       END::TEXT AS status,
       o.status::TEXT AS status_value
FROM "order" o
JOIN car_configuration c ON c.id = o.car_configuration_id
WHERE o.id = sqlc.arg(id) AND o.seller_id = sqlc.arg(seller_id);

-- Сохранять изменения может только продавец, оформивший заказ.
-- name: UpdateSellerOrder :execrows
UPDATE "order"
SET car_configuration_id = sqlc.arg(car_configuration_id),
    buyer_name = sqlc.arg(buyer_name),
    buyer_phone = sqlc.arg(buyer_phone),
    buyer_email = sqlc.narg(buyer_email),
    expected_delivery_on = sqlc.arg(expected_delivery_on)::DATE,
    status = sqlc.arg(status)::order_status
WHERE id = sqlc.arg(id) AND seller_id = sqlc.arg(seller_id);
