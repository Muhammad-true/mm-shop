# MM Shop - Production Release 
 
## 🚀 Быстрый старт 
 
1. Скопируйте папку `release` на сервер 
2. Подключитесь к серверу: `ssh root@000.00.00.000` 
3. Перейдите в папку: `cd release` 
4. Сделайте скрипт исполняемым: `chmod +x deploy-server.sh` 
5. Запустите деплой: `./deploy-server.sh` 
 

 
## 📊 Мониторинг 
- Проверить статус: `docker-compose -f docker-compose.release.yml ps` 
- Логи API: `docker-compose -f docker-compose.release.yml logs api` 
- Логи админки: `docker-compose -f docker-compose.release.yml logs admin` 

