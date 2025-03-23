#!/bin/sh

set -e

host="$1"
shift
port="$1"
shift

timeout=60

for i in $(seq $timeout); do
  nc -z "$host" "$port" > /dev/null 2>&1 && echo "✅ $host:$port доступен!" && exec "$@"
  sleep 1
done

echo "❌ Ошибка: $host:$port не доступен спустя $timeout секунд."
exit 1
