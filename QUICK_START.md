# 🚀 Быстрый старт для тестирования

## ✅ База данных успешно сброшена!

### 📊 Что было сделано:
- ✅ Удалены все пользователи (кроме дефолтных)
- ✅ Удалены все товары и вариации
- ✅ Удалены все заказы
- ✅ Удалены все изображения из `images/variations/`
- ✅ API перезапущен

---

## 👤 Доступные пользователи:

### 1. Super Admin
- **Email:** `admin@mm.com`
- **Пароль:** `password`
- **Роль:** `super_admin`
- **Может:** ВСЁ

### 2. Shop Owner 1
- **Email:** `shop1@mm.com`
- **Пароль:** `password`
- **Роль:** `shop_owner`
- **Может:** управлять СВОИМИ товарами и заказами

---

## 🎯 Следующие шаги:

### 1. Откройте админку
```
http://localhost:3000
```

### 2. Войдите как Super Admin
- Email: `admin@mm.com`
- Password: `password`

### 3. Сделайте Hard Refresh
- **Windows:** Ctrl + Shift + R
- **Mac:** Cmd + Shift + R

### 4. Начните тестирование!

---

## 📋 Быстрая проверка:

### Dashboard
- ✅ Показывает 0 товаров, 0 пользователей (или 2), 0 заказов
- ✅ Роль в консоли: `🎭 Нормализованная роль: super_admin`

### Товары
- ✅ Пустой список (можно создавать новые)

### Заказы
- ✅ Пустой список

### Категории
- ✅ Показываются дефолтные категории (5 шт.)

### Пользователи
- ✅ Показываются 2 пользователя (admin, shopowner)

### Роли
- ✅ Показываются 4 роли (super_admin, admin, shop_owner, user)

---

## 🔄 Если нужно сбросить снова:

```bash
# 1. Очистить БД
type reset-db-simple.sql | docker exec -i mm-postgres-dev psql -U mm_user -d mm_shop_dev

# 2. Удалить изображения
Remove-Item images\variations\*.jpg -Force -ErrorAction SilentlyContinue
Remove-Item images\variations\*.jpeg -Force -ErrorAction SilentlyContinue
Remove-Item images\variations\*.JPG -Force -ErrorAction SilentlyContinue
Remove-Item images\variations\*.JPEG -Force -ErrorAction SilentlyContinue
Remove-Item images\variations\*.png -Force -ErrorAction SilentlyContinue
Remove-Item images\variations\*.PNG -Force -ErrorAction SilentlyContinue

# 3. Перезапустить API
docker-compose restart api

# 4. Изменить роль admin на super_admin
docker exec mm-postgres-dev psql -U mm_user -d mm_shop_dev -c "UPDATE users SET role_id = (SELECT id FROM roles WHERE name = 'super_admin') WHERE email = 'admin@mm.com';"
```

---

## 🐛 Если что-то не работает:

### 1. Проверьте логи API
```bash
docker logs mm-api-dev --tail 50
```

### 2. Проверьте консоль браузера (F12)

### 3. Сделайте Hard Refresh (Ctrl+Shift+R)

### 4. Перезапустите контейнеры
```bash
docker-compose restart admin api
```

---

## 🎉 Готово к тестированию!

**Удачи! 🚀**

