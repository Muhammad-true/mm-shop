# 🚀 Быстрый старт: Настройка Cloudinary

## На Cloudinary (5 минут)

1. **Зарегистрируйтесь:** https://cloudinary.com/users/register/free

2. **Получите данные из Dashboard:**
   - Откройте: https://cloudinary.com/console
   - Скопируйте:
     - **Cloud name** (например: `driokajen`)
     - **API Key**
     - **API Secret**

3. **Создайте Upload Preset:**
   - Settings → Upload → Add upload preset
   - Название: `mm-shop-products`
   - Signing mode: **Unsigned**
   - Folder: `variations`
   - Format: `jpg`
   - Transformation: `w_1200,h_1200,c_fit,b_white,q_auto:good,fl_auto`
   - Сохраните

## На сервере (2 минуты)

1. **Отредактируйте `.env.production`:**
   ```bash
   cd ~/mm-shop/release
   nano .env.production
   ```

2. **Добавьте переменные:**
   ```bash
   USE_CLOUDINARY=true
   CLOUDINARY_CLOUD_NAME=ваш-cloud-name
   CLOUDINARY_API_KEY=ваш-api-key
   CLOUDINARY_API_SECRET=ваш-api-secret
   CLOUDINARY_UPLOAD_PRESET=mm-shop-products
   CLOUDINARY_REMOVE_BACKGROUND=false
   ```

3. **Перезапустите API:**
   ```bash
   docker compose -f docker-compose.release.yml restart api
   ```

4. **Проверьте логи:**
   ```bash
   docker compose -f docker-compose.release.yml logs -f api
   ```

5. **Загрузите тестовое фото:**
   - Через админ-панель загрузите фото товара
   - В логах должно быть: `☁️ Обработка изображения товара через Cloudinary...`
   - В ответе будет URL от Cloudinary

## ✅ Готово!

Теперь все фото товаров будут:
- ✅ Правильно поворачиваться (EXIF ориентация)
- ✅ Автоматически обрабатываться
- ✅ Храниться в Cloudinary с CDN

## 📖 Подробная инструкция

См. [CLOUDINARY_SETUP.md](./CLOUDINARY_SETUP.md)

