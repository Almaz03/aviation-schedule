#!/bin/bash

# Проверка наличия утилит
command -v psql >/dev/null 2>&1 || { echo >&2 "Ошибка: psql не установлен"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo >&2 "Ошибка: curl не установлен"; exit 1; }

# Очистка базы данных
echo "Очистка таблиц в PostgreSQL..."
psql -U your_postgres_user -d your_database_name <<EOF
TRUNCATE TABLE flights RESTART IDENTITY CASCADE;
TRUNCATE TABLE users RESTART IDENTITY CASCADE;
EOF

if [ $? -ne 0 ]; then
    echo "Ошибка при очистке базы данных"
    exit 1
fi

# Регистрация администратора
echo "Регистрация администратора..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  http://localhost:8082/register \
  -H "Content-Type: application/json" \
  -H "API-Key: khyWYbSHGjxUd98J2BwR4fNPrpgXv6ztZVmDAELqCs7Kc" \
  -d '{
    "username": "admin",
    "password": "admin",
    "role": "admin"
  }')

if [ "$response" -eq 200 ] || [ "$response" -eq 201 ]; then
    echo "Администратор успешно зарегистрирован"
else
    echo "Ошибка регистрации администратора. Код ответа: $response"
    exit 1
fi

echo "Скрипт выполнен успешно"
exit 0