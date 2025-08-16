## Delayed Notifier — отложенные уведомления

Сервис принимает заявки на отправку уведомлений в будущем, хранит их в Redis, публикует в RabbitMQ в нужный момент и доставляет (например, в Telegram). При ошибках отправка повторяется по экспоненциальной задержке. Есть простой веб‑интерфейс для ручного тестирования.

### Возможности

- Создание уведомления с датой/временем отправки: `POST /notify`
- Получение статуса: `GET /notify/{id}`
- Отмена: `DELETE /notify/{id}`
- UI на `static/index.html`
- Долгосрочное планирование (дни/недели) — за счёт Redis ZSET
- Повторы с экспоненциальной задержкой

### Архитектура (кратко)

- HTTP API: `wbf/ginext`
- Хранилище/планировщик: Redis ZSET (`notify:due`, `notify:retry`)
- Очередь: RabbitMQ (паблишер/консюмер)
- Доставщик: Telegram (можно расширять)
- Повторы: короткие in‑process через `github.com/kxddry/wbf/retry`, долгие — через Redis `notify:retry`

### Требования

- Docker/Docker Compose (для Redis, RabbitMQ)
- Go 1.24

### Быстрый старт

1) Запуск инфраструктуры

```bash
cd 1_delayed_notifier
docker compose up -d
# Redis: localhost:6379, RabbitMQ: localhost:5672, UI RabbitMQ: http://localhost:15672 (guest/guest)
```

2) Укажите токен Telegram (обязательно для реальной доставки)

Вариант A (предпочтительно): через переменные окружения/.env (файл .env подхватывается автоматически):

```bash
echo 'TELEGRAM_API_TOKEN=...' > .env
```

Вариант B (локально): в `config.yaml` поле `telegram.bot_token`

3) Запустите сервис

```bash
go run ./cmd/notifier
```

4) Откройте UI

```bash
open http://localhost:8080
```

### Конфигурация

Файл `config.yaml`:

- `server.host`, `server.port`, `server.static_dir`
- `redis.host`, `redis.port`, `redis.password`, `redis.db`
- `rabbitmq.host`, `rabbitmq.port`, `rabbitmq.username`, `rabbitmq.password`, `rabbitmq.queue_name`
- `telegram.bot_token` (может быть пустым, в проде используйте env `TELEGRAM_API_TOKEN`)

### API (пример)

- Создать уведомление (если `send_at` не указан — отправка сразу):

```bash
curl -X POST http://localhost:8080/notify \
  -H 'Content-Type: application/json' \
  -d '{
    "channel": "telegram",
    "recipient": "123456789",
    "message": "Привет!",
    "send_at": "2025-12-31T23:59:00Z"
  }'
```

Для канала `telegram` проверяется получатель: ровно 9 цифр. При несоответствии вернётся `400`.

- Статус:

```bash
curl http://localhost:8080/notify/<id>
```

- Отмена:

```bash
curl -X DELETE http://localhost:8080/notify/<id>
```

### Повторы и долгие задержки

- Короткие попытки в обработчике доставки: `retry.Do` со стратегией (3 попытки, delay 10ms, backoff x2)
- При неудаче — запись в Redis ZSET `notify:retry` со временем следующей попытки (экспоненциальная задержка до 6ч)
- Планировщик каждые ~1с вынимает due‑элементы из `notify:due` и `notify:retry` и публикует в RabbitMQ

### UI

- Доступен на `http://localhost:8080`
- Форма создания уведомления и проверка статуса/отмена

### Завершение работы (graceful shutdown)

- При SIGINT/SIGTERM сервис:
  - останавливает HTTP‑сервер c таймаутом
  - сигнализирует воркерам завершиться
  - закрывает подключения к RabbitMQ и Redis

### Тесты

```bash
cd 1_delayed_notifier
go test ./...
```

Покрыты кейсы неуспешной доставки, некорректного payload, отмены, валидации Telegram‑получателя, поведения планировщика и др.
