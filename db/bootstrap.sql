\getenv db_user PGUSER
\getenv db_name PGDATABASE
\getenv db_password PGPASSWORD

CREATE ROLE :"role" LOGIN PASSWORD :'db_password';
CREATE DATABASE :"db_name" OWNER :"db_user";
