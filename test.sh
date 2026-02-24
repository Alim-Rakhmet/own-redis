#!/bin/bash

HOST="127.0.0.1"
PORT="8080"

# Цвета для красивого вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Запуск автоматических тестов Own-Redis ===${NC}"
echo "Убедитесь, что ваш сервер запущен в другом окне: ./own-redis --port 8080"
echo "---------------------------------------------------"

# Функция для отправки запроса и проверки ответа
test_cmd() {
    local desc="$1"
    local cmd="$2"
    local expected="$3"
    local is_error_check="$4"

    # Отправляем команду через nc по UDP с таймаутом ожидания ответа
    # tr -d '\r\n' убирает переносы строк для удобного сравнения
    local result=$(echo "$cmd" | nc -u -w 1 $HOST $PORT 2>/dev/null | tr -d '\r\n')

    if [ "$is_error_check" == "true" ]; then
        # Если мы ждем ошибку, просто проверяем, содержит ли ответ слово error или ERR
        if [[ "$result" == *"error"* ]] || [[ "$result" == *"ERR"* ]]; then
            echo -e "[${GREEN}PASS${NC}] $desc"
        else
            echo -e "[${RED}FAIL${NC}] $desc (Ожидалась ошибка, получено: '$result')"
        fi
    else
        if [ "$result" == "$expected" ]; then
            echo -e "[${GREEN}PASS${NC}] $desc"
        else
            echo -e "[${RED}FAIL${NC}] $desc (Ожидалось: '$expected', Получено: '$result')"
        fi
    fi
}

# --- ГРУППА 1: Базовая проверка связи ---
test_cmd "1. Проверка команды PING" "PING" "PONG"
test_cmd "2. Нечувствительность к регистру (pInG)" "pInG" "PONG"

# --- ГРУППА 2: Базовые операции SET и GET ---
test_cmd "3. Стандартная запись ключа" "SET mykey 123" "OK"
test_cmd "4. Стандартное чтение ключа" "GET mykey" "123"
test_cmd "5. Чтение несуществующего ключа" "GET unknown_key" "(nil)"

# --- ГРУППА 3: Сложные значения и перезапись ---
test_cmd "6. SET значения из нескольких слов" "SET greeting Hello World and OpenAI" "OK"
test_cmd "7. GET значения с пробелами" "GET greeting" "Hello World and OpenAI"
test_cmd "8. Перезапись существующего значения" "SET mykey 456" "OK"
test_cmd "8.1 Проверка перезаписанного значения" "GET mykey" "456"

# --- ГРУППА 4: Обработка ошибок ---
test_cmd "9. Неизвестная команда" "DELETE mykey" "" "true"
test_cmd "10. Недостаточно аргументов для SET" "SET onlykey" "" "true"
test_cmd "11. Слишком много аргументов для GET" "GET key1 key2" "" "true"

# --- ГРУППА 5: Тестирование времени жизни (PX) ---
echo -e "${YELLOW}--- Запуск тестов со временем (PX) ---${NC}"
test_cmd "12. Установка ключа с PX 2000ms" "SET tempkey temp_val PX 2000" "OK"
test_cmd "13. Проверка ключа ДО истечения PX" "GET tempkey" "temp_val"

echo "Ждем 2.5 секунды, чтобы ключ протух..."
sleep 2.5
test_cmd "14. Проверка ключа ПОСЛЕ истечения PX" "GET tempkey" "(nil)"

echo "---------------------------------------------------"
echo -e "${YELLOW}=== Запуск Стресс-теста (Concurrency / Data Races) ===${NC}"
echo "Отправляем 100 одновременных запросов (SET и GET в фоне)..."

# Запускаем цикл, который отправляет запросы в фоне (&)
for i in {1..50}; do
    # Устанавливаем ключи
    echo "SET stress$i value$i" | nc -u -w 1 $HOST $PORT > /dev/null 2>&1 &
    # Пытаемся читать ключи одновременно с записью
    echo "GET stress$i" | nc -u -w 1 $HOST $PORT > /dev/null 2>&1 &
    # Пингуем параллельно
    echo "PING" | nc -u -w 1 $HOST $PORT > /dev/null 2>&1 &
done

# Ждем, пока все фоновые процессы завершатся
wait

# Проверяем, жив ли сервер (если упал из-за Data Race, PING не пройдет)
stress_result=$(echo "PING" | nc -u -w 1 $HOST $PORT 2>/dev/null | tr -d '\r\n')
if [ "$stress_result" == "PONG" ]; then
    echo -e "[${GREEN}PASS${NC}] 15. Сервер выдержал конкурентные запросы и не упал!"
else
    echo -e "[${RED}FAIL${NC}] 15. Сервер не отвечает после стресс-теста. Проверьте терминал сервера на наличие паник (panic: concurrent map read and map write)."
fi

echo -e "${YELLOW}=== Тестирование завершено ===${NC}"