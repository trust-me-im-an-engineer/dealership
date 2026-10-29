# Автосалон

Лабораторная работа № 10 по системному программному обеспечению, вариант 7.

Согласованный стек: PostgreSQL, Go, драйвер `pgx`. Интерфейс открывается в браузере; программа на Go реализует прикладную логику и обращается к PostgreSQL напрямую, без ODBC.

Приложение предусматривает режимы администратора и продавца. Администратор управляет данными об автомобилях, комплектациях и сроках поставки. Продавец заключает договоры покупки и заказывает автомобили нужной комплектации.

## Порядок запуска

Все команды выполняются из корня репозитория. В качестве агрегатора команд используется `just` (современный аналог make).

При первом запуске:

1. Создайте `.env` по образцу `.env.example` и задайте пароль PostgreSQL в
   `PGPASSWORD`. Значения `PGUSER` и `PGDATABASE` определяют имя создаваемых
   пользователя и базы данных.
2. Запустите службу PostgreSQL.
3. Создайте пользователя и базу командой `just db-init`.
4. Примените схему командой `just migrate-up`.
5. Загрузите тестовые данные командой `just db-seed`. Для входа используйте
   `admin` или `seller` с паролем `password`.
6. Запустите приложение командой `just run` и откройте
   <http://localhost:8080/login> (или адрес из `HTTP_ADDRESS`).

При последующих запусках достаточно запустить PostgreSQL и выполнить
`just run`. Повторно выполнять `db-init` и `db-seed` не нужно. Новые миграции
применяются командой `just migrate-up`.

## Работа с репозиторием

Для работы с репозиторием нужны:

- [Go 1.26](https://go.dev/doc/install);
- [just](https://just.systems/man/en/installation.html);
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) с поддержкой PostgreSQL;
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html);
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/local/);
- [Node.js и npm](https://nodejs.org/en/download) для форматирования HTML-шаблонов через Prettier;
- [PostgreSQL](https://www.postgresql.org/download/) — локально или на доступном сервере.

Перед запуском `just fmt` или `just lint` установите зависимости форматтера командой `npm ci`.
