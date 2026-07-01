# Quiz Platform

Веб-приложение для проведения квизов в реальном времени. Организаторы создают квизы и управляют сессиями, участники подключаются по коду комнаты и отвечают на вопросы вживую.

## Стек технологий

| Слой | Технология |
|---|---|
| Frontend | Next.js 14 (App Router), React 18, Tailwind CSS |
| Backend | Go (Fiber v2), WebSocket (gofiber/websocket) |
| База данных | PostgreSQL + sqlc |
| Аутентификация | JWT (access + refresh токены) |
| Хранилище файлов | Vercel Blob |

## Возможности

**Организатор**
- Регистрация и вход (email / пароль)
- Создание квиза: название, категория, сложность, теги, обложка
- Добавление вопросов: текст или изображение, одиночный / множественный выбор, таймер и баллы на вопрос
- Запуск сессии — генерация 6-значного кода комнаты
- Управление ходом квиза в реальном времени: переключение вопросов, просмотр лидерборда
- Личный кабинет: список квизов, история сессий, результаты

**Участник**
- Регистрация и вход
- Подключение к активному квизу по коду комнаты
- Ответы на вопросы в режиме реального времени
- Просмотр лидерборда по завершении квиза

## Запуск

### Требования
- Node.js 18+
- Go 1.22+
- PostgreSQL

### Backend

```bash
cd backend
```

Создайте файл `backend/.env`:

```env
DATABASE_URL=postgresql://user:password@localhost:5432/pulse_db
JWT_SECRET=your-secret-key
PORT=8080
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
```

Запустить (миграции применяются автоматически при старте):

```bash
go run ./cmd/server/main.go
```

### Frontend

```bash
npm install
```

Создайте файл `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

```bash
npm run dev
```

Приложение будет доступно по адресу [http://localhost:3000](http://localhost:3000).

## Структура проекта

```
.
├── backend/
│   ├── cmd/server/        # Точка входа, роуты
│   ├── db/
│   │   ├── migrations/    # SQL-миграции (golang-migrate)
│   │   ├── query/         # SQL-запросы (sqlc)
│   │   └── sqlc/          # Сгенерированный Go-код
│   ├── internal/
│   │   ├── auth/          # JWT manager
│   │   ├── handler/       # HTTP-хендлеры + DTO
│   │   ├── middleware/     # Auth, RequireRole
│   │   ├── repository/    # Store (обёртка над sqlc)
│   │   ├── service/       # Бизнес-логика
│   │   └── ws/            # WebSocket hub и клиенты
│   └── pkg/config/        # Конфигурация из .env
└── src/
    ├── app/               # Next.js App Router
    │   ├── (auth)/        # Вход, регистрация
    │   ├── dashboard/     # Личный кабинет
    │   ├── quiz/          # Создание, редактирование, запуск
    │   ├── play/[code]/   # Участие в квизе
    │   └── results/       # Результаты сессии
    ├── context/           # AuthProvider (JWT)
    └── lib/               # apiFetch, WebSocket client (ws.ts)
```

## WebSocket события

| Событие | Направление | Описание |
|---|---|---|
| `join-room` | клиент → сервер | Участник входит в сессию |
| `organizer-join` | организатор → сервер | Организатор подключается к комнате |
| `start-quiz` | организатор → сервер | Начало квиза |
| `next-question` | организатор → сервер | Следующий вопрос |
| `submit-answer` | участник → сервер | Отправка ответа |
| `question-ended` | сервер → всем | Время вышло, показ правильных ответов |
| `score-update` | сервер → всем | Обновление лидерборда |
| `quiz-finished` | сервер → всем | Квиз завершён, финальные результаты |

## Скрипты

```bash
# Frontend
npm run dev      # Режим разработки
npm run build    # Production-сборка
npm run lint     # ESLint

# Backend
go run ./cmd/server/main.go   # Запуск
go build ./...                # Проверка сборки
```
