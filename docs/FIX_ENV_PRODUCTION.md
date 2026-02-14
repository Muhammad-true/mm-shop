# 🔧 Исправление проблемы с .env.production в Git

## Проблема
Файл `.env.production` отслеживается Git, хотя должен быть в `.gitignore`.

## Решение на сервере

### Шаг 1: Временно сохранить файл и сделать pull

```bash
cd ~/mm-shop/release

# Сохранить файл во временное место
cp .env.production .env.production.backup

# Удалить файл (чтобы Git мог сделать pull)
rm .env.production

# Теперь можно сделать pull
git pull

# Вернуть файл обратно
cp .env.production.backup .env.production

# Удалить backup (опционально)
rm .env.production.backup
```

### Шаг 2: Проверить, что файл в .gitignore

```bash
# Проверить, что .env.production в .gitignore
grep "\.env.production" .gitignore

# Должно показать:
# .env.production  # Production конфигурация (секреты, не коммитим!)
```

### Шаг 3: Проверить статус Git

```bash
# Проверить статус - .env.production не должен показываться
git status

# Если файл все еще показывается, удалите его из индекса
git rm --cached .env.production 2>/dev/null || true
```

## Если файл все еще отслеживается

Если после всех действий файл все еще показывается в `git status`:

```bash
# Полностью удалить из Git истории (но оставить на диске)
git rm --cached .env.production

# Закоммитить удаление из индекса
git commit -m "Remove .env.production from Git tracking"

# Отправить изменения
git push origin main
```

## Добавление переменных Cloudinary

После того, как файл перестанет отслеживаться Git:

```bash
# Отредактировать файл
nano .env.production

# Добавить переменные:
USE_CLOUDINARY=true
CLOUDINARY_CLOUD_NAME=ваш_cloud_name
CLOUDINARY_API_KEY=ваш_api_key
CLOUDINARY_API_SECRET=ваш_api_secret
CLOUDINARY_UPLOAD_PRESET=ваш_preset_name
CLOUDINARY_REMOVE_BACKGROUND=false

# Сохранить (Ctrl+O, Enter, Ctrl+X)
```

## Перезапуск контейнера

```bash
# Перезапустить контейнер, чтобы загрузить новые переменные
docker compose -f docker-compose.release.yml restart api

# Проверить переменные в контейнере
docker compose -f docker-compose.release.yml exec api printenv | grep CLOUDINARY
```

## Быстрое решение (одной командой)

```bash
cd ~/mm-shop/release && \
cp .env.production .env.production.backup && \
rm .env.production && \
git pull && \
cp .env.production.backup .env.production && \
rm .env.production.backup && \
git status
```
