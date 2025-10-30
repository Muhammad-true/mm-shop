-- Создание супер-админа
-- Пароль: admin123

-- 1. Создаем роль super_admin, если её нет
INSERT INTO roles (id, name, display_name, description, permissions, is_active, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    'super_admin',
    'Супер Администратор',
    'Полный доступ ко всей системе',
    '{"dashboard": true, "users": true, "products": true, "categories": true, "orders": true, "roles": true, "settings": true}'::jsonb,
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE name = 'super_admin');

-- 2. Создаем пользователя супер-админа
INSERT INTO users (id, name, email, password, phone, is_active, role_id, created_at, updated_at)
SELECT 
    gen_random_uuid(),
    'Super Admin',
    'admin@mm.com',
    '$2a$10$iwztid0QXT6v7XW59FhV6.55MXDrORiLsiY1anvdY9DKcpf/R4xJq', -- admin123
    '927781020',
    true,
    (SELECT id FROM roles WHERE name = 'super_admin'),
    NOW(),
    NOW()
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@mm.com');

-- 3. Проверяем результат
SELECT u.id, u.name, u.email, u.phone, r.name as role
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.email = 'admin@mm.com';

