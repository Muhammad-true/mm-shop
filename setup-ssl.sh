#!/bin/bash
# Скрипт для создания самоподписанного SSL сертификата
# Использование: ./setup-ssl.sh [domain_or_ip]

set -e

DOMAIN="${1:-159.89.99.252}"
SSL_DIR="./ssl"

echo "🔒 Создание SSL сертификата для $DOMAIN..."

# Создаем директорию для сертификатов
mkdir -p "$SSL_DIR"

# Генерируем самоподписанный сертификат
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "$SSL_DIR/key.pem" \
  -out "$SSL_DIR/cert.pem" \
  -subj "/C=RU/ST=State/L=City/O=MM Shop/CN=$DOMAIN"

# Устанавливаем правильные права
chmod 600 "$SSL_DIR/key.pem"
chmod 644 "$SSL_DIR/cert.pem"

echo "✅ SSL сертификат создан:"
echo "   Certificate: $SSL_DIR/cert.pem"
echo "   Private Key: $SSL_DIR/key.pem"
echo ""
echo "⚠️  Это самоподписанный сертификат. Для продакшена используйте Let's Encrypt!"
echo ""
echo "📝 Для использования Let's Encrypt:"
echo "   1. Установите certbot: sudo apt-get install certbot"
echo "   2. Получите сертификат: sudo certbot certonly --standalone -d your-domain.com"
echo "   3. Обновите docker-compose.release.yml: замените volume на /etc/letsencrypt:/etc/letsencrypt:ro"
echo "   4. Обновите nginx.production.conf: укажите пути к Let's Encrypt сертификатам"

