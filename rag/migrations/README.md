# Миграции базы данных

Этот каталог содержит миграции базы данных для проекта RAG, использующие [golang-migrate](https://github.com/golang-migrate/migrate).

## Структура файлов

Миграции должны следовать формату: `{номер}_{описание}.{up|down}.sql`

- `{номер}` - порядковый номер миграции (например, 000001, 000002)
- `{описание}` - краткое описание изменений
- `{up|down}` - направление миграции:
  - `up` - применение миграции
  - `down` - откат миграции

## Использование

### Запуск с Docker Compose

```bash
# Запуск PostgreSQL и применение всех миграций
docker-compose up -d

# Только запуск PostgreSQL без миграций
docker-compose up -d postgres

# Применение миграций вручную
docker-compose run --rm migrate
```

### Ручное управление миграциями

```bash
# Применить все миграции
docker run --rm -v $(pwd)/migrations:/migrations \
  --network rag_default \
  migrate/migrate:latest \
  -path /migrations \
  -database "postgres://rag_user:rag_password@rag_postgres:5432/rag_db?sslmode=disable" \
  up

# Откатить последнюю миграцию
docker run --rm -v $(pwd)/migrations:/migrations \
  --network rag_default \
  migrate/migrate:latest \
  -path /migrations \
  -database "postgres://rag_user:rag_password@rag_postgres:5432/rag_db?sslmode=disable" \
  down 1

# Проверить статус миграций
docker run --rm -v $(pwd)/migrations:/migrations \
  --network rag_default \
  migrate/migrate:latest \
  -path /migrations \
  -database "postgres://rag_user:rag_password@rag_postgres:5432/rag_db?sslmode=disable" \
  version
```

## Создание новой миграции

1. Создайте два файла:
   - `{номер}_{описание}.up.sql` - для применения изменений
   - `{номер}_{описание}.down.sql` - для отката изменений

2. Убедитесь, что номер миграции больше всех существующих

3. Протестируйте миграцию локально перед коммитом

## Подключение к базе данных

```bash
# Через psql в контейнере
docker exec -it rag_postgres psql -U rag_user -d rag_db

# Через внешний клиент
psql -h localhost -p 5432 -U rag_user -d rag_db
```

## Переменные окружения

- `POSTGRES_DB`: rag_db
- `POSTGRES_USER`: rag_user  
- `POSTGRES_PASSWORD`: rag_password
- `DATABASE_URL`: postgres://rag_user:rag_password@postgres:5432/rag_db?sslmode=disable 