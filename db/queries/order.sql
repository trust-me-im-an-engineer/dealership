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
