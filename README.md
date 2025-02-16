# Merch Store

## Магазин мерча

В Авито существует внутренний магазин мерча, где сотрудники могут приобретать товары за монеты (coin). Каждому новому сотруднику выделяется 1000 монет, которые можно использовать для покупки товаров. Кроме того, монеты можно передавать другим сотрудникам в знак благодарности или как подарок.

## Развертывание при помощи Docker

### 1. Клонировать репозиторий

```sh
git clone https://github.com/MelDxKviel/merch-store.git
```

### 2. Настроить переменные окружения при необходимости

В файле `docker-compose.yml`

Переменные окружения базы данных
```yml
...
environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: merchstoredb
...
```
Переменные окружения api
```yml
...
environment:
      DATABASE_URL: "postgres://postgres:postgres@db:5432/merchstoredb?sslmode=disable"
      JWT_SECRET: "secret"
...
```

### 3. Запустить приложение с помощью `docker compose`

```sh
docker compose up --build
```

---

По умолчанию сервис доступен по http://localhost:8080