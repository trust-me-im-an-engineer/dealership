-- psql получает имена и пароль из стандартных переменных окружения PostgreSQL.
\getenv db_user PGUSER
\getenv db_name PGDATABASE
\getenv db_password PGPASSWORD

-- Создание пользователя и базы данных с именами из переменных окружения.
CREATE ROLE :"db_user" LOGIN PASSWORD :'db_password';
CREATE DATABASE :"db_name" OWNER :"db_user";
