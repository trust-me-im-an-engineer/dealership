\getenv db_name DB_NAME
\getenv db_user DB_USER
\getenv db_password DB_PASSWORD

CREATE ROLE :"db_user" LOGIN PASSWORD :'db_password';
CREATE DATABASE :"db_name" OWNER :"db_user";
