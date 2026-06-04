-- Включаем расширение для генерации UUID, если его еще нет в базе
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE, -- Опциональное поле, поэтому без NOT NULL
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создаем индекс по user_id, чтобы база данных мгновенно находила подписки конкретного пользователя!
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
