#!/bin/bash

set -e
set -u

function create_user_and_database() {
	local database=$1
	echo "  Creating user and database '$database'"
	psql --username "$POSTGRES_USER" --dbname "postgres" <<-EOSQL
	    CREATE DATABASE $database;
	    GRANT ALL PRIVILEGES ON DATABASE $database TO $POSTGRES_USER;
EOSQL
}

if [ -n "$POSTGRES_MULTIPLE_DATABASES" ]; then
	echo "Multiple database creation requested: $POSTGRES_MULTIPLE_DATABASES"
	for db in $(echo $POSTGRES_MULTIPLE_DATABASES | tr ',' ' '); do
		create_user_and_database $db
	done
	echo "Multiple databases created"
fi

# Создаем служебные таблицы в основной базе данных
echo "Creating service tables in mm_shop_dev database..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "mm_shop_dev" <<-EOSQL
    -- Таблица для логирования
    CREATE TABLE IF NOT EXISTS service_logs (
        id SERIAL PRIMARY KEY,
        level VARCHAR(10) NOT NULL,
        message TEXT NOT NULL,
        context JSONB,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    -- Таблица для метрик
    CREATE TABLE IF NOT EXISTS service_metrics (
        id SERIAL PRIMARY KEY,
        metric_name VARCHAR(100) NOT NULL,
        metric_value DECIMAL(15,2) NOT NULL,
        tags JSONB,
        recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    -- Таблица для кэша
    CREATE TABLE IF NOT EXISTS service_cache (
        cache_key VARCHAR(255) PRIMARY KEY,
        cache_value TEXT NOT NULL,
        expires_at TIMESTAMP WITH TIME ZONE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    -- Индексы для оптимизации
    CREATE INDEX IF NOT EXISTS idx_service_logs_level ON service_logs(level);
    CREATE INDEX IF NOT EXISTS idx_service_logs_created_at ON service_logs(created_at);
    CREATE INDEX IF NOT EXISTS idx_service_metrics_name ON service_metrics(metric_name);
    CREATE INDEX IF NOT EXISTS idx_service_metrics_recorded_at ON service_metrics(recorded_at);
    CREATE INDEX IF NOT EXISTS idx_service_cache_expires_at ON service_cache(expires_at);

    -- Создаем пользователя для служебных операций
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'service_user') THEN
            CREATE ROLE service_user WITH LOGIN PASSWORD 'service_password';
        END IF;
    END
    \$\$;

    -- Даем права служебному пользователю
    GRANT SELECT, INSERT, UPDATE, DELETE ON service_logs TO service_user;
    GRANT SELECT, INSERT, UPDATE, DELETE ON service_metrics TO service_user;
    GRANT SELECT, INSERT, UPDATE, DELETE ON service_cache TO service_user;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO service_user;

    -- Создаем представления для удобства
    CREATE OR REPLACE VIEW recent_logs AS
    SELECT level, message, context, created_at
    FROM service_logs
    WHERE created_at > NOW() - INTERVAL '24 hours'
    ORDER BY created_at DESC;

    CREATE OR REPLACE VIEW system_metrics AS
    SELECT metric_name, AVG(metric_value) as avg_value, 
           MAX(metric_value) as max_value, 
           MIN(metric_value) as min_value,
           COUNT(*) as count
    FROM service_metrics
    WHERE recorded_at > NOW() - INTERVAL '1 hour'
    GROUP BY metric_name;

EOSQL

echo "Service tables and views created successfully!" 