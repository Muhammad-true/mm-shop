# MM Shop - Production Release 
 
## 🚀 Быстрый старт 
 
1. Скопируйте папку `release` на сервер 
2. Подключитесь к серверу: `ssh root@159.89.99.252` 
3. Перейдите в папку: `cd release` 
4. Сделайте скрипт исполняемым: `chmod +x deploy-server.sh` 
5. Запустите деплой: `./deploy-server.sh` 
 
## 🌐 Доступные сервисы 
- **Админ панель:** http://159.89.99.252 
- **API:** http://159.89.99.252:8080 
- **PgAdmin:** http://159.89.99.252:8081 
 
## 📊 Мониторинг 
- Проверить статус: `docker-compose -f docker-compose.release.yml ps` 
- Логи API: `docker-compose -f docker-compose.release.yml logs api` 
- Логи админки: `docker-compose -f docker-compose.release.yml logs admin` 
