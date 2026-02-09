#!/bin/bash

# Скрипт для получения API токена
# Использование: ./get_api_token.sh

API_BASE_URL="${API_BASE_URL:-https://api.libiss.com/api/v1}"

echo "🔑 Получение API токена"
echo ""

# Запрашиваем данные
read -p "Телефон: " PHONE
read -sp "Пароль: " PASSWORD
echo ""

# Выполняем запрос
echo "📤 Отправка запроса..."
RESPONSE=$(curl -s -X POST "$API_BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"phone\": \"$PHONE\",
    \"password\": \"$PASSWORD\"
  }")

# Парсим токен
TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Ошибка получения токена"
    echo "Ответ сервера:"
    echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"
    exit 1
fi

echo "✅ Токен получен:"
echo ""
echo "$TOKEN"
echo ""
echo "📋 Используй этот токен в переменной окружения:"
echo "export API_TOKEN=\"$TOKEN\""
echo ""

