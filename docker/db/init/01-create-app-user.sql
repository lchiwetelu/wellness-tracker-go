CREATE USER wellness_app WITH PASSWORD 'wellness_password';

GRANT ALL PRIVILEGES ON DATABASE wellness TO wellness_app;

GRANT ALL ON SCHEMA public TO wellness_app;