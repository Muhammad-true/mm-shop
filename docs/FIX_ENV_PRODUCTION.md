# 🔧 Исправление проблемы с .env.production в Git

## Проблема
Файл `.env.production` отслеживается Git, хотя должен быть в `.gitignore`.

## Решение на сервере

### Шаг 1: Убрать файл из staging area и индекса Git

```bash
cd ~/mm-shop/release

# Убрать из staging area
git restore --staged .env.production

# Убрать из индекса Git (но оставить на диске)
git rm --cached .env.production

# Проверить статус
git status
```

### Шаг 2: Получить обновленный .gitignore

```bash
# Теперь можно сделать pull
git pull
```

### Шаг 3: Проверить, что файл игнорируется

```bash
# Проверить статус - .env.production не должен показываться
git status

# Проверить, что файл в .gitignore
grep "\.env.production" .gitignore
```

## Если файл все еще отслеживается

Если после `git rm --cached` файл все еще показывается:

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

