\getenv db_user PGUSER
\getenv db_name PGDATABASE
\getenv db_password PGPASSWORD

CREATE ROLE :"db_user" LOGIN PASSWORD :'db_password';
CREATE DATABASE :"db_name" OWNER :"db_user";
