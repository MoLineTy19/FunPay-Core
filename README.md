[![Go Version](https://img.shields.io/github/go-mod/go-version/MoLineTy19/FunPay-Core)](https://github.com/MoLineTy19/FunPay-Core)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![golangci-lint](https://github.com/MoLineTy19/FunPay-Core/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/MoLineTy19/FunPay-Core/actions/workflows/golangci-lint.yml)
[![Latest Release](https://img.shields.io/github/v/release/MoLineTy19/FunPay-Core)](https://github.com/MoLineTy19/FunPay-Core/releases/latest)

# FunPay Core

Go-движок для автоматизации FunPay. Берёт на себя всё общение с FunPay: авторизация, polling событий, заказы, чаты, лоты, баланс. Снаружи отдаёт состояние через REST на localhost. Внутри нет базы, нет UI, нет шаблонов выдачи. Состояние живёт в памяти, пока движок запущен.

Один бинарь, один аккаунт продавца, один порт. Бот или дашборд поверх делает остальное.

---

## Зачем это

Автоматизация продавца на FunPay обычно скатывается в один из двух вариантов. Либо толстый скрипт, где авторизация, бизнес-логика и UI смешаны в кашу. Либо чужой закрытый бот, которому отдаёшь golden key, не понимая, что он внутри делает.

FunPay Core режет задачу иначе. Движок знает только FunPay: как залогиниться, как получить события, как выдать заказ, как ответить в чат. Всё остальное (шаблоны авто-выдачи, операторские уведомления, FSM, история) живёт у бота. Между ними один мост: REST-контракт на localhost. Ядро никогда не стучит в бот, бот всегда инициатор.

Первый готовый сценарий: авто-выдача текстового товара по новому заказу с уведомлением продавца в Telegram. Архитектура умеет и больше: управление лотами, чаты, возвраты, баланс.

---

## Архитектура

Три слоя, у каждого строгая зона ответственности. Зависимости идут только вниз.

```
internal/fp/        знает FunPay: URL, форматы, HTML, auth, throttle
internal/engine/    превращает сырые ответы в поток событий с монотонным ID
internal/rest/      контракт для бота: REST + long-poll, не знает про FunPay
cmd/engine/         точка сборки, читает .env, связывает слои адаптерами
```

Если FunPay поменяет API, правки останутся внутри `internal/fp/`. Если захочется другой бот, веб-дашборд или CLI-клиент вместо Telegram, движок трогать не нужно. Контракт REST развязывает технологии.

### Слоистый контракт на уровне ошибок

REST-хендлеры не знают FunPay напрямую. Они работают с интерфейсами (`OrderLister`, `OrderGetter`, `ChatMessager`, `OrderRefunder`) и ловят экспортированные sentinel-ошибки из `fp` через `errors.Is`. Так `ErrAuthLost` становится HTTP 503, `ErrOrderNotFound` становится 404, всё остальное становится 500. Маппинг локализован в одном месте, типизировать ошибки можно без знания внутренностей FunPay.

Полный список sentinel-ошибок в `internal/fp`:

- `ErrAuthLost` (golden_seal истёк или отсутствует) → 503 `auth_lost`
- `ErrOrderNotFound` → 404 `order_not_found`
- `ErrOfferNotFound` → 404 `offer_not_found`
- `ErrChatNotFound` → 404 `chat_not_found`

Формат тела ошибки одинаковый для всех эндпоинтов:

```json
{
  "error": {
    "code": "auth_lost",
    "message": "auth lost: golden_seal expired or missing",
    "retryable": false
  }
}
```

`retryable: true` означает внутреннюю ошибку (500), которую стоит повторить. Остальные коды детерминированы, повтор того же запроса ничего не изменит.

### Эскроу-модель

FunPay работает по предоплате в момент заказа. Покупатель жмёт «Купить», оплачивает, деньги уходят в эскроу. Заказ появляется у продавца уже оплаченным. Триггер авто-выдачи это `order.new`, отдельного «order.paid» нет. К моменту, когда FunPay отдаёт заказ как новый, он уже оплачен.

---

## Запуск

```bash
cp .env.example .env
# заполнить FP_* и ENGINE_TOKEN

go run ./cmd/engine/
```

Флаг `--debug` поднимает уровень логирования до debug.

Сборка и проверки:

```bash
go build ./...
go test ./...
go vet ./...
```

Тесты покрывают парсеры на реальных образцах HTML, pure-функцию `diffOrderSnapshots` по всем веткам, error-mapping и REST-хендлеры.

### Docker

В репозитории есть многоэтапный `Dockerfile` (builder на `golang:1.25-alpine`, runtime на distroless `nonroot`) и `docker-compose.yml`.

Движок по дизайну читает `.env` с диска — и при старте (`godotenv.Load()`), и при resume после `auth_lost` (`godotenv.Read()`). Поэтому в Docker `.env` **монтируется как volume, а не запекается в образ** и не передаётся через `environment:`. Так resume-флоу работает как есть: оператор правит `.env` на хосте → `POST /control/resume` → движок перечитывает файл.

```bash
cp .env.example .env
# заполнить FP_* и ENGINE_TOKEN
# ВАЖНО для Docker: ENGINE_LISTEN=0.0.0.0:8731 (127.0.0.1 внутри контейнера недоступен снаружи)

docker compose up -d --build
docker compose logs -f engine
```

Порт `8731` проброшен на `127.0.0.1:8731` хоста (доступен только локально). Для доступа с других машин замените маппинг в `docker-compose.yml` на `"8731:8731"`.

Сборка с версионированием (как в CI):
```bash
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg COMMIT=$(git rev-parse --short HEAD) \
  --build-arg DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t funpay-engine .
```

Обновить `FP_GOLDEN_SEAL` после `auth_lost` без пересборки образа:
```bash
$EDITOR .env                    # правим seal на хосте
curl -X POST http://127.0.0.1:8731/control/resume \
  -H "X-Engine-Token: $ENGINE_TOKEN"
```

`.env` не попадает в образ (см. `.dockerignore`); секреты остаются только в смонтированном файле на хосте.

### `.env`

```env
FP_GOLDEN_KEY=
FP_PHPSESSID=
FP_GOLDEN_SEAL=
FP_USER_ID=
FP_CSRF_TOKEN=

ENGINE_TOKEN=
ENGINE_LISTEN=127.0.0.1:8731
```

`FP_*` достаются из DevTools после ручного логина на funpay.com: `golden_key`, `phpsessid`, `golden_seal` из Application → Cookies; `user_id` и `csrf_token` из Network-вкладки на runner-запросе.

`ENGINE_TOKEN` это общий секрет движка и бота. Сгенерируйте случайную строку и используйте то же значение в `.env` бота. Пустой токен фатален: движок не стартует без защиты.

`ENGINE_LISTEN` это адрес привязки. По умолчанию `127.0.0.1:8731`, то есть только localhost.

---

## Авторизация и «человеческий темп»

Вход: golden key плюс корректные заголовки браузера. Никакого TLS-fingerprinting, никакого headless-браузера.

Запросы к FunPay идут через `Throttler`: минимальная задержка плюс случайный jitter, чтобы трафик не выглядел роботизированным. Параметры задаются в `NewClient` (минимум 800 мс, jitter 600 мс по умолчанию).

`golden_seal` истекает. Когда движок ловит `ErrAuthLost`, он перестаёт polling, эмитит `engine.status` со state `auth_lost` и ждёт. Оператор обновляет seal в `.env` и шлёт `POST /control/resume`. Движок перечитывает `.env`, обновляет auth в памяти, реинициализирует runner и продолжает с того же `last_event_id`. Если seal в `.env` не изменился, движок остаётся в паузе и ждёт реального обновления.

---

## REST API

REST слушает на localhost (по умолчанию `127.0.0.1:8731`). Каждый запрос требует заголовок `X-Engine-Token: <ENGINE_TOKEN>`. Без него 401, с пустым токеном движок вообще не стартует.

Все тела запросов и ответов в JSON, кодировка UTF-8.

### Эндпоинты

| Метод | Путь | Назначение |
|------|------|------------|
| GET | `/health` | uptime + размер буфера событий + state (`healthy` / `auth_lost`) |
| POST | `/events/poll` | события из буфера, long-poll |
| GET | `/account` | userId, login, balance, loadedAt |
| POST | `/offers` | создать лот |
| PATCH | `/offers/{node}/{offer}` | редактировать лот |
| DELETE | `/offers/{node}/{offer}` | удалить лот |
| GET | `/offers/{node}` | список лотов раздела |
| GET | `/offers/form?node=X` | схема формы + список серверов |
| GET | `/orders` | список продаж |
| GET | `/orders/{id}` | детали заказа |
| POST | `/orders/{id}/refund` | вернуть заказ |
| POST | `/chats/{id}/messages` | ответить покупателю в чат |
| POST | `/control/resume` | восстановить polling после auth_lost |

Про пути. `{node}` в чатах это идентификатор вида `users-{buyerId}-{sellerId}`, а не числовой chatId. Параметр пути у чата называется `{id}`, но фактически туда кладётся node. `{id}` заказа это код вроде `WMBY8JNK`. `{node}` у лотов это ID раздела FunPay, `{offer}` это ID лота внутри раздела.

### `POST /events/poll`

Long-poll событий. Тело:

```json
{ "since": 42, "wait": 15 }
```

`since` это последний обработанный `eventId` у бота (с первого раза `0`). `wait` это сколько секунд держать соединение, если событий нет, максимум 30. Ответ:

```json
{
  "events": [ { "eventId": 43, "type": "order.new", "at": "...", "payload": { ... } } ],
  "nextEventId": 43
}
```

Если за `wait` ничего не пришло, вернётся пустой массив `events` без `nextEventId`. Бот хранит свой `last_event_id` и в следующий раз шлёт `since` равным `nextEventId` из последнего ответа.

Если `since` указывает на событие, уже вытесненное из буфера (буфер in-memory с TTL), движок отвечает 409 `cursor_too_old`. В этом случае бот должен сбросить состояние и прочитать актуальный снимок через `/orders`, `/account` и т.п.

### `POST /offers`

Создать лот. Тело:

```json
{
  "nodeId": "80",
  "serverId": "7",
  "fields": { "summary": { "ru": "300 голды за 10 минут" }, "delivery": { "ru": "автоматически" } },
  "price": "150.00",
  "amount": 0,
  "active": true
}
```

Обязательны `nodeId`, `serverId`, непустой `fields.summary` и `price >= 0`. `amount` и `active` опциональны. Валидация: отсутствие одного из обязательных полей даёт 400 `bad_request`. Ответ 201:

```json
{ "nodeId": "80", "offerId": "12345", "created": true, "url": "https://funpay.com/..." }
```

### `PATCH /offers/{node}/{offer}`

Частичное редактирование. В теле передаёте только то, что меняете:

```json
{ "price": "160.00" }
```

Доступные поля: `fields`, `price`, `amount`, `active`. Запрос без единого поля даёт 400: «nothing to update». `price` не может быть отрицательным. Ответ:

```json
{ "nodeId": "80", "offerId": "12345", "updated": true, "url": "https://funpay.com/..." }
```

### `DELETE /offers/{node}/{offer}`

Удаление. Ответ:

```json
{ "nodeId": "80", "offerId": "12345", "deleted": true }
```

### `GET /offers/{node}`

Список лотов в разделе. Пустой список возвращается как `[]`, не `null`:

```json
{
  "nodeId": "80",
  "offers": [
    { "offerId": "12345", "summary": "300 голды", "server": "Alpha", "amount": "50", "price": "150.00" }
  ]
}
```

### `GET /offers/form?node=X`

Схема формы создания лота для раздела `node`: какие поля ввода ожидает FunPay (с типами) и какие серверы доступны для выбора.

```json
{
  "nodeId": "80",
  "serverId": "7",
  "fields": [ { "id": "summary", "type": 1 } ],
  "servers": [ { "id": "7", "name": "Alpha" } ]
}
```

Эта схема нужна, чтобы бот заранее знал, какие `fields` и `serverId` подсунуть в `POST /offers`, не разбирая HTML самостоятельно.

### `GET /orders`

Список продаж:

```json
{
  "orders": [
    {
      "id": "WMBY8JNK",
      "status": "active",
      "buyerName": "player1",
      "summary": "300 голды",
      "price": "150.00",
      "chatId": "users-100-200"
    }
  ]
}
```

### `GET /orders/{id}`

Детали одного заказа:

```json
{
  "id": "WMBY8JNK",
  "offerId": "12345",
  "nodeId": "80",
  "buyerId": 100,
  "buyerName": "player1",
  "amount": "150.00",
  "currency": "RUB",
  "status": "active",
  "createdAt": "2024-01-15T12:00:00Z",
  "chatId": "users-100-200"
}
```

### `POST /orders/{id}/refund`

Вернуть заказ. Тела нет. Ответ:

```json
{ "ok": true, "orderId": "WMBY8JNK" }
```

### `POST /chats/{id}/messages`

Ответить покупателю в чат. Параметр `{id}` здесь это node чата (`users-{buyerId}-{sellerId}`), который можно взять из `chatId` заказа или события. Тело:

```json
{ "text": "Ваш заказ готов: ABC-KEY-123" }
```

Пустой `text` даёт 400. Ответ:

```json
{ "ok": true, "messageId": "88123456" }
```

### `POST /control/resume`

Восстановить polling после `auth_lost`. Тела нет. Перед вызовом оператор должен обновить `FP_GOLDEN_SEAL` в `.env`. Движок сравнивает новый seal с тем, что в памяти: если совпадает (не обновлён) или пуст, остаётся в паузе. Если отличается, перечитывает весь auth, реинициализирует runner и продолжает опрос.

---

## События

События это способ связи между ядром и ботом. Движок складывает их в in-memory буфер, бот забирает через `POST /events/poll`. Каждое событие несёт монотонный `eventId`, тип и payload.

```
order.new          новый заказ (уже оплачен, эскроу)
order.completed    покупатель подтвердил, деньги перешли продавцу
order.cancelled    возврат или спор закрыт не в пользу продавца
chat.message       новое сообщение покупателя
offer.changed      состояние лота изменилось
account.balance    изменился баланс
engine.status      движок перешёл в auth_lost или восстановился
```

Бот хранит свой `last_event_id`. После перезапуска движка бот просто снова читает события с этой позиции, ничего не теряется, потому что персистентное состояние живёт у бота, а не в ядре. У буфера есть TTL, старые события вытесняются, поэтому при долгом простое бот может получить 409 `cursor_too_old` и должен пересобрать снимок через REST.

### Как детектируется новый заказ

Runner получает счётчики `orders_counters = {buyer, seller}`. Это только числа, без ID заказа. Когда `seller > 0` или `buyer > 0`, движок тянет свежий список `/orders/trade` и сравнивает с предыдущим снимком через pure-функцию `diffOrderSnapshots`. Появился новый ID, эмитится `order.new`. Статус сменился на completed или cancelled, эмитится соответствующее событие. Всё остальное (включая уход заказа из списка) игнорируется. Снимок заменяется целиком. На старте baseline грузится без эмита, иначе все существующие заказы стали бы «новыми».

---

## Что внутри слоёв

`internal/fp/` единственный слой, знающий FunPay.

- `client.go` — HTTP-клиент, auth, throttle, csrf-токен
- `runner.go` — stateful polling loop: tags, bookmarks, снимок заказов, diff
- `runner_init.go` — инициализация runner (baseline без эмита)
- `sales.go` — парсинг HTML списка `/orders/trade`
- `order_detail.go` — парсинг HTML детальной `/orders/{id}/`
- `chat_send.go` — SendMessage через `POST /runner/` с `action=chat_message`
- `chat_node_state.go` — состояние чата (last_message + tag) перед отправкой
- `order_refund.go` — возврат через `POST /orders/refund`
- `offer_form.go`, `my_offers.go`, `offer_edit.go`, `offer_save.go` — CRUD лотов
- `account.go` — профиль и баланс продавца
- `throttle.go` — throttler с jitter
- `types.go` — типы и sentinel-ошибки

`internal/engine/` — тонкий декоратор.

- `buffer.go` — in-memory буфер событий с TTL и eviction
- `decoder.go` — `WrapEvents`: превращает `RunnerEvents` в типизированные события
- `events.go` — константы типов событий

`internal/rest/` — сервер на стандартном `net/http` (Go 1.22+ pattern routing).

- Хендлеры делегируют интерфейсам, не знают FunPay
- Error-mapping через `errors.Is` на sentinel-ошибки fp
- DTO: `AccountSnapshot`, `OrderListItem`, `OrderDetail`, `OfferForm` и т.д.

`cmd/engine/main.go` — точка сборки: читает `.env`, создаёт клиент и runner, связывает слои через адаптеры, крутит poll-цикл и loop обновления аккаунта.

---

## Disclaimer

Проект не аффилирован с FunPay. Автоматизация может нарушать правила площадки. Используйте на свой риск.
