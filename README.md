Blog Platform API

Расширенная система управления блогом на Go. Реализована с использованием чистой архитектуры (Clean Architecture), конкурентного планировщика публикаций и полного покрытия unit-тестами.

## Быстрый старт

### Требования

- Go 1.21+
- Docker & Docker Compose

### Установка и запуск

1. **Клонируйте репозиторий:**

```bash
git clone https://github.com/workjung719/blog-platform.git
cd blog-platform
```

2. Запустите базу данных PostgreSQL:**
  

```bash
docker-compose up -d
```

*Проверьте статус контейнера командой `docker ps`. Он должен быть `Up (healthy)`.*

3. **Настройте окружение:**

Создайте файл `.env` на основе примера:

```bash
cp .env.example .env
```

**Важно:** Измените значение `JWT_SECRET` в файле `.env` на случайную строку длиной минимум 32 символа. Файл `.env` добавлен в `.gitignore` и не попадает в репозиторий.

4. **Установите зависимости и запустите сервер:**

```bash
go mod tidy
go run cmd/api/main.go
```

Сервер запустится на `http://localhost:8080`. При старте автоматически применяются миграции БД и запускается фоновый планировщик публикаций.

## Архитектура

Проект построен по принципу разделения ответственности (Separation of Concerns):

cmd/api/main.go # Точка входа, инициализация зависимостей, маршрутизация
internal/
├── config/ # Загрузка конфигурации из переменных окружения
├── model/ # Структуры данных (DTO) и методы валидации
├── repository/ # Слой доступа к данным (SQL запросы через sqlx)
├── service/ # Бизнес-логика (валидация прав, хеширование, JWT)
├── handler/ # HTTP обработчики (парсинг запросов, формирование ответов)
├── middleware/ # Промежуточные слои (Auth, Logging, Panic Recovery)
└── scheduler/ # Планировщик отложенных публикаций (Worker Pool)
migrations/ # SQL скрипты миграций (встроены через go:embed)

**Ключевые особенности:**

- **DIP (Dependency Inversion):** Сервисы зависят от интерфейсов репозиториев, а не от конкретных реализаций. Это позволяет легко мокать БД в тестах.
- **Единый формат ошибок:** Централизованная обработка исключений через пакет `respond`.
- **Graceful Shutdown:** Корректное завершение работы HTTP-сервера и фонового планировщика при получении сигнала SIGINT/SIGTERM.

## Безопасность

- **Хеширование паролей:** Используется алгоритм `bcrypt` (DefaultCost). Пароли никогда не хранятся в открытом виде.
- **JWT Аутентификация:** Токены подписываются строго алгоритмом `HS256`. В middleware проверяется корректность подписи и срок действия токена.
- **Защита от SQL-инъекций:** Все запросы к базе данных параметризованы (`$1, $2...`).
- **Валидация входных данных:** Строгая проверка email (RFC), длины username/password на уровне модели перед обращением к БД.
- **Конфиденциальность:** Поле `password_hash` исключено из JSON-сериализации (`json:"-"`). Секреты вынесены в `.env`, который игнорируется Git.

## API Endpoints

### Публичные эндпоинты


| Метод | Путь | Описание |
| --- | --- | --- |
| GET | `/api/health` | Проверка состояния сервиса и подключения к БД |
| POST | `/api/register` | Регистрация нового пользователя |
| POST | `/api/login` | Авторизация, возвращает JWT токен |
| GET | `/api/posts` | Список опубликованных постов (пагинация: `?limit=10&offset=0`) |
| GET | `/api/posts/{id}` | Получить конкретный пост |
| GET | `/api/users/{userId}/posts` | Получить все посты конкретного автора |
| GET | `/api/posts/{postId}/comments` | Получить комментарии к посту |

### Защищенные эндпоинты (требуется Header: `Authorization: Bearer <token>`)


| Метод | Путь | Описание | Права |
| --- | --- | --- | --- |
| POST | `/api/posts` | Создать пост (можно указать `publish_at` для отложенной публикации) | Любой авторизованный |
| PUT | `/api/posts/{id}` | Обновить пост | Только автор поста (иначе 403 Forbidden) |
| DELETE | `/api/posts/{id}` | Удалить пост | Только автор поста (иначе 403 Forbidden) |
| POST | `/api/posts/{postId}/comments` | Добавить комментарий | Любой авторизованный |
| PUT | `/api/comments/{commentId}` |  Обновить комментарий | Только автор комментария (иначе 403) |
| DELETE | `/api/comments/{commentId}` |  Удалить комментарий | Только автор комментария (иначе 403) |

## Планировщик отложенных публикаций

Система поддерживает создание постов с future timestamp (`publish_at`).

- При создании такого поста ему присваивается статус `draft`.
  
- Фоновая горутина (**Scheduler**) каждые N секунд (настраивается в `.env`) сканирует таблицу `posts`.
  
- Находит записи со статусом `draft`, где `publish_at <= NOW()`.
  
- Публикует их параллельно через **Worker Pool** (количество воркеров настраивается).
  
- Логирует процесс публикации.
  
- Останавливается корректно при завершении приложения (Context Cancellation).
  

## Тестирование

Написаны comprehensive unit-тесты для бизнес-логики без обращения к реальной БД (использование моков):

- **Models:** Валидация email, пароля, username.
- **Services:** Логика создания постов (статусы draft/published), проверка прав доступа (owner check), пагинация.
- **Middleware:** Валидация JWT токенов (валидные, битые, чужие подписи).
- **Scheduler:** Логика обработки due-постов, изоляция ошибок.

Запуск тестов:

```bash
go test ./... -v
```

Покрытие кода:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Примеры запросов (Curl)

Ниже приведены рабочие примеры взаимодействия с API. 
**Важно:** Перед выполнением защищенных запросов необходимо получить JWT токен через `/api/login`.

### 1. Проверка здоровья сервиса (Health Check)

```bash
curl http://localhost:8080/api/health
```

*Ожидаемый ответ: `{"data":{"status":"ok"}}`*

### 2. Регистрация пользователя

```bash
curl -X POST http://localhost:8080/api/register \
-H "Content-Type: application/json" \
-d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "SecurePass123!"
}'
```

*Ответ: `201 Created`, содержит поле `"token"`.*

### 3. Вход (Получение токена)

```bash
curl -X POST http://localhost:8080/api/login \
-H "Content-Type: application/json" \
-d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
}'
```

*Cохраните значение поля `"token"` из ответа для дальнейших шагов.*

### 4. Создание поста (Обычный / Published)

```bash
curl -X POST http://localhost:8080/api/posts \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "title": "Мой первый пост",
    "content": "Привет, мир Go!"
}'
```

*Статус созданного поста будет `published`.*

### 5. Создание отложенного поста (Draft)

```bash
# Указываем время публикации в будущем (ISO 8601)
curl -X POST http://localhost:8080/api/posts \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "title": "Завтрашний пост",
    "content": "Этот текст появится автоматически завтра.",
    "publish_at": "2026-10-03T12:00:00Z"
}'
```

*Статус созданного поста будет `draft`. Планировщик опубликует его по наступлении времени.*

### 6. Получение списка постов (Пагинация)

```bash
curl "http://localhost:8080/api/posts?limit=10&offset=0"
```

*Возвращает только опубликованные посты.*

### 7. Получение конкретного поста по ID

*(Замените `1` на ID существующего поста)*

```bash
curl -X PUT http://localhost:8080/api/posts/1 \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "title": "Обновленный заголовок",
    "content": "Новый контент поста."
}'
```

### 8. Обновление поста (Только автор)

*(Замените `1` на ID вашего поста)*

```bash
curl -X PUT http://localhost:8080/api/posts/1 \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "title": "Обновленный заголовок",
    "content": "Новый контент поста."
}'
```

###

### 9. Удаление поста (Только автор)

```bash
curl -X DELETE http://localhost:8080/api/posts/1 \
-H "Authorization: Bearer <YOUR_TOKEN>"
```

*Ответ: `200 OK` или `204 No Content`.*

### 10. Получение постов конкретного автора

*(Замените `1` на ID пользователя)*

```bash
curl http://localhost:8080/api/users/1/posts
```

### 11. Добавление комментария к посту

*(Замените `postId` на ID существующего поста)*

```bash
curl -X POST http://localhost:8080/api/posts/1/comments \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "content": "Отличная статья!"
}'
```

### 12. Получение комментариев к посту

```bash
curl http://localhost:8080/api/posts/1/comments
```

### 13. Обновление комментария (Только автор комментария)

*(Замените `commentId` на ID комментария)*

```bash
curl -X PUT http://localhost:8080/api/comments/1 \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <YOUR_TOKEN>" \
-d '{
    "content": "Исправленный комментарий"
}'
```

## Технологии

- **Language:** Go 1.21+
- **Router:** [chi/v5](https://github.com/go-chi/chi)
- **Database Driver:** [lib/pq](https://github.com/lib/pq), [sqlx](https://github.com/jmoiron/sqlx)
- **Auth:** [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt), [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Config:** [godotenv](https://github.com/joho/godotenv)
- **Logging:** Standard library `log/slog`
- **Infrastructure:** Docker, Docker Compose, PostgreSQL 15
