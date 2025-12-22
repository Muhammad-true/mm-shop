-- Добавление поля lemonsqueezy_variant_id в таблицу subscription_plans
-- Это поле хранит ID варианта продукта из Lemon Squeezy для каждого плана подписки

ALTER TABLE subscription_plans 
ADD COLUMN IF NOT EXISTS lemonsqueezy_variant_id VARCHAR(255);

-- Создаем индекс для быстрого поиска по lemonsqueezy_variant_id
CREATE INDEX IF NOT EXISTS idx_subscription_plans_lemonsqueezy_variant_id 
ON subscription_plans(lemonsqueezy_variant_id);

-- Комментарий к полю
COMMENT ON COLUMN subscription_plans.lemonsqueezy_variant_id IS 'ID варианта продукта из Lemon Squeezy для этого плана подписки';

