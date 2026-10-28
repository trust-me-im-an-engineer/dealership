-- Расширение создаёт bcrypt-хеши паролей для тестовых учётных записей.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Администратор и продавец входят с одинаковым тестовым паролем.
INSERT INTO "user" (role, username, password_hash)
VALUES ('admin', 'admin', crypt('password', gen_salt('bf'))),
       ('seller', 'seller', crypt('password', gen_salt('bf')));

-- Тестовый каталог автомобилей и комплектаций.
INSERT INTO car_model (name, year)
VALUES ('Lada Vesta', 2025),
       ('Geely Coolray', 2025);

INSERT INTO car_configuration (name, car_model, price, delivery_days, description)
VALUES ('Comfort', 'Lada Vesta', 1500000.00, 14, 'Базовая комплектация'),
       ('Techno', 'Lada Vesta', 1800000.00, 21, 'Расширенная комплектация'),
       ('Flagship', 'Geely Coolray', 2700000.00, 30, 'Максимальная комплектация');

-- Заказ в работе и уже доставленный автомобиль.
INSERT INTO "order" (car_configuration_id, seller_id, buyer_name, buyer_phone, buyer_email,
                     expected_delivery_on, status)
SELECT c.id, u.id, 'Иван Петров', '+7 900 000-00-01', 'ivan@example.com',
       CURRENT_DATE + c.delivery_days, 'placed'
FROM car_configuration AS c CROSS JOIN "user" AS u
WHERE c.car_model = 'Lada Vesta' AND c.name = 'Comfort' AND u.username = 'seller';

INSERT INTO "order" (car_configuration_id, seller_id, buyer_name, buyer_phone, buyer_email,
                     ordered_at, expected_delivery_on, status)
SELECT c.id, u.id, 'Анна Смирнова', '+7 900 000-00-02', 'anna@example.com',
       NOW() - INTERVAL '40 days', CURRENT_DATE - 10, 'delivered'
FROM car_configuration AS c CROSS JOIN "user" AS u
WHERE c.car_model = 'Geely Coolray' AND c.name = 'Flagship' AND u.username = 'seller';
