# Автосалон

Лабораторная работа № 10 по системному программному обеспечению, вариант 7.

Согласованный стек: PostgreSQL, Go, драйвер `pgx`. Интерфейс открывается в браузере; программа на Go реализует прикладную логику и обращается к PostgreSQL напрямую, без ODBC.

Приложение предусматривает режимы администратора и продавца. Администратор управляет данными об автомобилях, комплектациях и сроках поставки. Продавец заключает договоры покупки и заказывает автомобили нужной комплектации.

## Работа с репозиторием

Для работы с репозиторием нужны:

- [Go 1.26](https://go.dev/doc/install);
- [just](https://just.systems/man/en/installation.html);
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) с поддержкой PostgreSQL;
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html);
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/local/);
- [PostgreSQL](https://www.postgresql.org/download/) — локально или на доступном сервере.
