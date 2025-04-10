#!/bin/bash

# ========== НАСТРОЙКИ ==========
CONCURRENCY=${1:-500}
DURATION=${2:-60}
WAVES=${3:-5}
INTERVAL=${4:-5}
PORT=9100
BINARY=mock-users
LOG_FILE=mock-users.log

if lsof -i :9100 > /dev/null; then
  echo "⛔ Порт 9100 уже занят. Завершаю старый процесс..."
  fuser -k 9100/tcp
  sleep 1
fi

# ========== СБОРКА ==========
echo "🚀 Собираем $BINARY..."
go build -o $BINARY main.go metrics.go load.go

# ========== ЗАПУСК ==========
echo "⚡ Запускаем $BINARY с нагрузкой:"
echo "   - $CONCURRENCY горутин / endpoint"
echo "   - $DURATION сек работы"
echo "   - $WAVES волн с интервалом $INTERVAL сек"
echo ""

CONCURRENCY=$CONCURRENCY DURATION=$DURATION WAVES=$WAVES INTERVAL=$INTERVAL ./$BINARY > $LOG_FILE 2>&1 &
PID=$!

# ========== ТАЙМЕР ==========
for ((i=0; i<$DURATION; i++)); do
  echo -ne "⏱ Прошло: $i сек / $DURATION сек\r"
  sleep 1
done

# ========== ЗАВЕРШЕНИЕ ==========
echo ""
echo "🛑 Останавливаем mock-users (PID $PID)..."
kill -9 $PID

echo "✅ Завершено. Логи сохранены в $LOG_FILE"
echo "📊 Открой Grafana: http://localhost:3000"
