# 🔗 URL Shortener

Полнофункциональный сервис для сокращения ссылок с аналитикой переходов, построенный на Go с использованием Clean Architecture.

## Возможности

- ✅ Создание коротких ссылок (POST /shorten)
- ✅ Поддержка кастомных имен для ссылок
- ✅ Автоматический редирект на оригинальный URL
- ✅ Сбор аналитики переходов (IP, User-Agent, время)
- ✅ Агрегация статистики по дням/месяцам
- ✅ Простой веб-интерфейс для тестирования

## Технологии

- **Go 1.23+** - основной язык
- **PostgreSQL 15+** - хранение данных
- **wbf/dbpg** - библиотека от Wildberries для работы с БД (retry, master-slave)
- **net/http** - стандартная библиотека Go для HTTP
- **Docker & Docker Compose** - контейнеризация

## Быстрый старт

### 1. Настройка окружения

Скопируйте `.env.example` в `.env` и настройте параметры:

```bash
cp .env.example .env
```

### 2. Запуск через Docker Compose

```bash
docker-compose up -d
```

Сервис будет доступен на `http://localhost:8080`

### 3. Запуск локально

Требуется PostgreSQL. Создайте базу данных:

```sql
CREATE DATABASE urlshortener;
```

Установите зависимости и запустите:

```bash
go mod download
go run cmd/main.go
```

## API Endpoints

### POST /shorten
Создание короткой ссылки

**Request:**
```json
{
  "url": "https://example.com/very/long/url",
  "custom_short": "my-link"  // опционально
}
```

**Response:**
```json
{
  "short_url": "my-link",
  "full_url": "https://example.com/very/long/url"
}
```

### GET /{short_url}
Редирект на оригинальный URL (автоматически записывает клик)

### GET /analytics/{short_url}
Получение статистики переходов

**Query параметры:**
- `start_date` - начальная дата (формат: YYYY-MM-DD)
- `end_date` - конечная дата (формат: YYYY-MM-DD)

**Response:**
```json
{
  "short_url": "my-link",
  "full_url": "https://example.com/very/long/url",
  "total_clicks": 42,
  "clicks_by_day": [
    {
      "date": "2026-02-09T00:00:00Z",
      "count": 15
    }
  ]
}
```

### GET /analytics/{short_url}/user-agents
Статистика по User-Agent

**Response:**
```json
{
  "short_url": "my-link",
  "user_agent_stats": [
    {
      "user_agent": "Mozilla/5.0...",
      "count": 25
    }
  ]
}
```

## Веб-интерфейс

Откройте `http://localhost:8080` в браузере для доступа к UI:

- Форма создания коротких ссылок
- Просмотр аналитики по ссылкам
- Копирование ссылок в буфер обмена

## Архитектура

Проект следует Clean Architecture:

```
cmd/                    # Точка входа
internal/
  ├── domain/          # Доменные модели
  ├── port/            # Интерфейсы (порты)
  ├── usecases/        # Бизнес-логика
  ├── adapter/         # Адаптеры (репозитории)
  └── input/           # HTTP обработчики
pkg/                   # Общие пакеты
web/                   # Статические файлы UI
```

## База данных

### Таблица urls
- `id` - уникальный идентификатор
- `full_url` - оригинальный URL
- `short_url` - короткая ссылка
- `created_at` - время создания

### Таблица url_clicks
- `id` - уникальный идентификатор
- `short_url_id` - внешний ключ на urls
- `clicked_at` - время перехода
- `ip_address` - IP адрес
- `user_agent` - User-Agent браузера

## Примеры использования

### Создание ссылки с кастомным именем

```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com", "custom_short": "gh"}'
```

### Получение аналитики

```bash
curl "http://localhost:8080/analytics/gh?start_date=2026-01-01&end_date=2026-02-09"
```

## Разработка

### Структура проекта

- `net/http` - стандартная библиотека Go для HTTP сервера
- `wbf/dbpg` - для работы с PostgreSQL с retry-стратегией и master-slave
- Миграции применяются автоматически при старте
- Clean Architecture - разделение на слои (domain, ports, usecases, adapters)

### Добавление новых функций

1. Определите интерфейс в `internal/port/`
2. Реализуйте use case в `internal/usecases/`
3. Добавьте адаптер в `internal/adapter/`
4. Создайте HTTP handler в `internal/input/http/`

### Полезные документы

- [QUICKSTART.md](QUICKSTART.md) - быстрый старт
- [ARCHITECTURE.md](ARCHITECTURE.md) - подробное описание архитектуры
- [API_EXAMPLES.md](API_EXAMPLES.md) - примеры использования API

## Лицензия

MIT
