# Backend интернет-магазина

Go, Gin, PostgreSQL (database/sql и pgx) и Redis.

## Локальный запуск

1. Скопируйте `.env.example` в `.env` и задайте случайный `JWT_SECRET` длиной не менее 32 байт.
2. Запустите `docker compose up --build -d`. API доступно на `http://localhost:8080`, проверка готовности — `GET /health`.
3. Остановить контейнеры: `docker compose stop`. Данные PostgreSQL сохраняются в Docker volume.

Схема создаётся при запуске. Промокод `SAVE10` добавляется автоматически. Для назначения администратора сначала зарегистрируйте пользователя, затем выполните `docker compose exec postgres psql -U shop -d shop -c "UPDATE users SET role='admin' WHERE email='admin@example.com';"`, указав его email.

Переменные: `DATABASE_URL` — адрес PostgreSQL; `REDIS_URL` — адрес Redis; `JWT_SECRET` — секрет подписи; `HTTP_ADDR` — адрес HTTP-сервера. В Compose адреса базы и Redis заданы для внутренней сети контейнеров; учётные данные базы в Compose предназначены только для локальной разработки.

## API

- `POST /api/auth/register`, `POST /api/auth/login`: JSON с `email`, `password`; вход возвращает `token`.
- `GET /api/products`: параметры `page`, `limit`, `search`, `min_price`, `max_price`, `min_rating`, `sort_by` (`title` или `price`), `order` (`asc` или `desc`).
- Администратор: `POST /api/products`, `PUT /api/products/:id`, `DELETE /api/products/:id`. При создании и замене передаются `title`, `description`, `price`, `rating`, `images` (не менее трёх URL), `stock_quantity`.
- Авторизованный пользователь: `GET /api/cart`, `POST /api/cart/items` (`product_id`, `quantity`), `PATCH /api/cart/items/:id` (`quantity`), `DELETE /api/cart/items/:id`, `POST /api/cart/apply-promo` (`code`). ID элемента корзины возвращается при добавлении.

Защищённые запросы требуют заголовок `Authorization: Bearer <token>`. Каталог кэшируется в Redis на 90 секунд.

Проект пока не содержит OpenAPI, настройки CORS и облачного развёртывания. Использование Go по заданию необходимо согласовать с преподавателем.
