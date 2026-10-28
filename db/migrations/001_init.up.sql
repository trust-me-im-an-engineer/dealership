-- Возможные состояния заказа автомобиля.
CREATE TYPE order_status AS ENUM ('placed', 'delivered', 'cancelled');

-- Учётные записи пользователей.
CREATE TABLE "user"
(
    id            BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    role          TEXT NOT NULL CHECK (role IN ('admin', 'seller')),
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

-- Модель автомобиля и год её выпуска.
CREATE TABLE car_model
(
    name TEXT PRIMARY KEY,
    year INT NOT NULL
);

-- Доступные комплектации с ценой и сроком поставки.
CREATE TABLE car_configuration
(
    id            BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name          TEXT           NOT NULL,
    car_model     TEXT           NOT NULL REFERENCES car_model (name),
    price         NUMERIC(12, 2) NOT NULL CHECK (price > 0),
    delivery_days INT            NOT NULL CHECK (delivery_days >= 0),
    description   TEXT
);

-- Заказы покупателей, оформленные продавцами.
CREATE TABLE "order"
(
    id                   BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    car_configuration_id BIGINT       NOT NULL REFERENCES car_configuration (id),
    seller_id            BIGINT       NOT NULL REFERENCES "user" (id),
    buyer_name           TEXT         NOT NULL,
    buyer_phone          TEXT         NOT NULL,
    buyer_email          TEXT,
    ordered_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expected_delivery_on DATE         NOT NULL,
    status               order_status NOT NULL DEFAULT 'placed'
);
