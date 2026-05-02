# Процесс разработки

## Статус этапов

| Этап | Статус |
|------|--------|
| 1 — Инициализация | ✅ Закрыт |
| 2 — Авторизация и роли | ✅ Закрыт |
| 3 — Помещения и Бронирование | ✅ Закрыт |
| 4 — Frontend: Клиент (Студент) | ✅ Закрыт (ожидает проверки) |
| 5 — Frontend: Админ и Схемы | ⏳ Ожидает |
| 6 — Уведомления и Фоновые задачи | ⏳ Ожидает |
| 7 — Деплой | ⏳ Ожидает |

## Этап 4 — выполнен (2026-05-02)
Весь код Фронтенда Студента написан (Vanilla HTML/JS/CSS):
- **Дизайн**: Glassmorphism (blur, semi-transparent background, Aurora effect), темная тема.
- **TMA**: Интеграция с `Telegram.WebApp`, автоматическая отправка `init_data` и скрытие/отрисовка компонентов через SPA подход (переключение `.view`).
- **Бронирование**: 
  - Выбор коворкинга.
  - Указание времени через нативные селекторы `date`/`time` и выгрузка списка свободных/занятых мест.
  - Подтверждение брони через попап или Telegram Alerts.
- **История**: Просмотр своих броней и кнопка отмены (с `confirm`).
- **Тесты**: Подключен Jest и `jest-environment-jsdom`, написаны unit-тесты (`frontend/tests/app.test.js`) на класс `API` и функцию `switchView`.

### Команды для проверки:
```powershell
cd frontend
npm install
npm run test
npm run lint
```

### Команда для коммита:
```powershell
git add frontend/
git commit -m "feat(stage4): student TMA frontend with glassmorphism UI and jest tests"
```

---
Весь код Этапа 3 написан:
- **Миграции**: `003_global_settings`
- **Репозиторий**: CRUD комнат/мест с availability, ядро бронирования (транзакции, лимиты, overlap).
- **Бизнес-логика**: `internal/service/booking.go` с полным покрытием unit-тестами.
- **Хендлеры и Middleware**: Protected + Admin Only маршруты, полный CRUD.
- **Интеграционные тесты**: `api_stage3_test.go` покрывает весь флоу: бронирование, конфликты (409), лимиты (422), отмену, пагинацию, admin update.

### Команды для проверки:
```powershell
go test ./...
go test -tags=integration ./...
golangci-lint run ./...
```

### Команда для коммита:
```powershell
git add .
git commit -m "feat(stage3): rooms/workspaces CRUD, booking core with conflict detection and limits"
```

---

# Процесс разработки – Этап 1 (Bootstrap)

## Что уже выполнено

| Шаг | Описание | Файл/Компонент |
|-----|----------|----------------|
| 1️⃣  | Инициализация репозитория Git | `git init` (создан `.git` каталог) |
| 2️⃣  | Базовый README с обзором проекта | `README.md` |
| 3️⃣  | .gitignore с типовыми записями для Go, Node, IDE, OS | `.gitignore` |
| 4️⃣  | Go‑модуль и зависимости (`sqlx`, `gorilla/mux`, `lib/pq`, `golang-migrate`) | `go.mod` |
| 5️⃣  | Точка входа HTTP‑сервера в `cmd/server/main.go` | `cmd/server/main.go` |
| 6️⃣  | Загрузчик конфигурации (порт, DSN) | `internal/config/config.go` |
| 7️⃣  | Базовый bootstrap‑слой приложения (создание DB, репозитория, роутера) | `internal/app/app.go` |
| 8️⃣  | Обёртка подключения к PostgreSQL через `sqlx` | `internal/db/postgres.go` |
| 9️⃣  | Модели данных: `Room`, `Workspace`, `User`, `Booking`, `BookingSettings` | `internal/models/models.go` |
| 🔟  | Репозиторий с заглушкой получения всех помещений | `internal/repository/repo.go` |
| 1️⃣1️⃣ | Роутер с эндпоинтом `/health` | `internal/router/router.go` |
| 1️⃣2️⃣ | Конфигурация golangci‑lint (включены основные линтеры) | `.golangci.yml` |
| 1️⃣3️⃣ | Ранее добавлен черновой блок команд в process (ниже сохранён как архив копирования) | см. блок «архив команд» ниже |
| 1️⃣4️⃣ | SQL‑миграции `001_initial` + программное применение + CLI | `migrations/*`, `internal/dbmigrations/`, `cmd/migrate/` |
| 1️⃣5️⃣ | Docker: `Dockerfile`, `docker-compose.yml` (`db`, `migrate`, `api`), `.dockerignore` | корень |
| 1️⃣6️⃣ | GitHub Actions: unit + golangci‑lint + integration + frontend ESLint | `.github/workflows/ci.yml` |
| 1️⃣7️⃣ | Тесты: unit `/health`, интеграция БД под `-tags=integration` (+ `RUN_INTEGRATION` / `CI`) | `internal/router/router_test.go`, `internal/db/integration_test.go` |
| 1️⃣8️⃣ | Frontend: TMA заготовка + ESLint 9 + Prettier | `frontend/` |
| 1️⃣9️⃣ | `README.md` дополнен командами; `.golangci.yml` без `depguard`; исключения revive/stylecheck по package‑комментариям | `README.md`, `.golangci.yml` |
| 2️⃣1️⃣ | Миграция `002_auth_seed`: колонка `password_hash`, seed‑администратор | `migrations/002_*` |
| 2️⃣2️⃣ | Валидация Telegram Web Apps `init_data` (HMAC), unit‑тест | `internal/auth/telegram.go`, `telegram_test.go`, `SignedInitDataForTests` (`//go:build integration`) |
| 2️⃣3️⃣ | JWT (`golang-jwt/jwt/v5`), сервис входа студента и админа | `internal/auth/*.go`, `JWT_SECRET`, `JWT_EXPIRES_HOURS`, `TELEGRAM_BOT_TOKEN` в `internal/config/` |
| 2️⃣4️⃣ | HTTP: `/api/auth/admin/login`, `/api/auth/student/telegram`, middleware Bearer, `/api/rooms` | `internal/handlers/*`, `internal/middleware/jwt.go`, `internal/router/router.go` |
| 2️⃣5️⃣ | Репозиторий: `GetAdminByEmail`, `UpsertStudentTelegramUser` | `internal/repository/users.go` |
| 2️⃣6️⃣ | Интеграционный сценарий API без циклических импортов | `internal/integration/api_auth_test.go`, `internal/testsupport/integration.go` |
| 2️⃣7️⃣ | Документация и compose: секрет JWT, переменная бота; CI env | `README.md`, `docker-compose.yml`, `.github/workflows/ci.yml` |

## Этап 1 — статус (2026-05-02)

**Git:** результат Этапа 1 зафиксирован **одним** коммитом на ветке `main` *(subject: `feat(stage1): Postgres migrations, migrate CLI, CI, and TMA scaffold`; актуальный хеш — `git log -1 --oneline`)*.

План из согласованного сообщения выполнен: база смоделирована миграциями, воспроизводимость — через `cmd/migrate` и сервис `migrate` в compose, CI включает линтеры и интеграционные тесты (Postgres‑сервис), фронт — продовый скелет без «отладочных» текстов Telegram (сообщение о выборе комнаты только на чистый браузер вне клиента Telegram).

Локально интеграционные тесты **skipped**, пока не заданы `RUN_INTEGRATION=1` или общий признак CI — это осознанно, чтобы `./...` без Docker не падал. Пакеты с покрытием: `internal/db`, `internal/integration` (API `httptest`), общий помощник — `internal/testsupport`.

## Этап 2 — статус (авторизация и роли, backend)

Закрывает формулировку Этапа 2 из `description.md`: студент проходит связку **регистрация+вход** через подписанный `init_data` Mini App (без пароля студента на сервере), администратор — по email/паролю из учётной записи, созданной миграцией; выдаётся JWT; защищённый эндпоинт списка помещений; интеграционные тесты на успешный логин и отказ при неверных данных; линтеры зелёные.

Дальнейший шаг по плану — **Этап 3**: CRUD помещений/мест, ядро бронирований, лимиты.

## Архив: команды‑черновик (до реализации, можно не копировать)
```bash
# 1️⃣ Миграции
mkdir -p migrations
cat > migrations/001_init.up.sql <<'EOF'
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);
CREATE TABLE workspaces (
    id SERIAL PRIMARY KEY,
    room_id INT REFERENCES rooms(id) ON DELETE CASCADE,
    name TEXT NOT NULL
);
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL CHECK (role IN ('student','admin')),
    email TEXT
);
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    workspace_id INT REFERENCES workspaces(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL
);
CREATE TABLE booking_settings (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    max_active INT NOT NULL DEFAULT 3
);
EOF

cat > migrations/001_init.down.sql <<'EOF'
DROP TABLE IF EXISTS booking_settings;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS rooms;
EOF

# 2️⃣ Docker‑compose
cat > docker-compose.yml <<'EOF'
version: "3.9"
services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: coworking_db
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
  api:
    build: .
    depends_on:
      - db
    environment:
      DB_DSN: "postgres://postgres:postgres@db:5432/coworking_db?sslmode=disable"
      PORT: "8080"
    ports:
      - "8080:8080"
volumes:
  pgdata:
EOF

# 3️⃣ CI‑pipeline (GitHub Actions)
mkdir -p .github/workflows
cat > .github/workflows/ci.yml <<'EOF'
name: CI
on: [push, pull_request]
jobs:
  build-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: coworking_db
        ports: [5432:5432]
        options: >-
          --health-cmd "pg_isready -U postgres"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Install dependencies
        run: go mod tidy
      - name: Run linters
        run: golangci-lint run
      - name: Run tests
        run: go test ./...
EOF
EOF

# 4️⃣ Тест здоровья
mkdir -p internal/app/test
cat > internal/app/test/health_test.go <<'EOF'
package app_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/example/coworking/internal/app"
    "github.com/example/coworking/internal/config"
    "github.com/example/coworking/internal/router"
    "github.com/example/coworking/internal/db"
    "github.com/example/coworking/internal/repository"
)

func TestHealthEndpoint(t *testing.T) {
    cfg, _ := config.LoadConfig()
    dbConn, _ := db.NewDB(cfg.DBDSN)
    repo := repository.NewRepo(dbConn)
    r := router.NewRouter(repo)
    a, _ := app.NewApp(cfg)
    a.Router().(*router.Router).Mux().Handle("/health", r.Mux().HandlerFunc)

    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    rr := httptest.NewRecorder()
    a.Router().ServeHTTP(rr, req)
    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200 OK, got %d", rr.Code)
    }
}
EOF

# 5️⃣ Frontend скелет
mkdir -p frontend
cat > frontend/index.html <<'EOF'
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Коворкинг – Mini App</title>
    <link rel="stylesheet" href="styles.css">
    <script src="https://telegram.org/js/telegram-web-app.js"></script>
    <script src="app.js" defer></script>
</head>
<body>
    <div id="app">Loading…</div>
</body>
</html>
EOF

cat > frontend/app.js <<'EOF'
window.addEventListener('load', () => {
    const tg = window.Telegram.WebApp;
    tg.ready();
    document.getElementById('app').textContent = 'Коворкинг Mini App готов!';
});
EOF

cat > frontend/styles.css <<'EOF'
:root {
    --bg: hsl(220, 10%, 10%);
    --fg: hsl(0, 0%, 95%);
    --accent: hsl(210, 80%, 55%);
}
html, body {margin:0;padding:0;height:100%;background:#111;color:var(--fg);font-family:'Inter',sans-serif;}
#app {display:flex;align-items:center;justify-content:center;height:100%;font-size:1.5rem;}
EOF

# 6️⃣ Запуск
docker compose up -d
```

*После выполнения всех команд проверьте работоспособность:* `curl http://localhost:8080/health` → должно вернуть `OK`.

> **Примечание:** в блоке CI выше есть опечатка — лишний `EOF` перед закрытием блока при копировании в терминал проверять вручную.

---

## Журнал 2026-05-02

### Источник ТЗ

- В корне есть **`description.md`** (полный план из 7 этапов). Отдельного файла **`descr.md`** нет — дальнейшая трактовка по `description.md`.

### Прогресс по этапу

- **Утром:** только анализ и письменный план (до ответа «ОК»).
- **После «ОК»:** реализованы миграции, `docker-compose`/Dockerfile, CI, тесты, frontend/Lint конфигурация и обновление `README` (см. таблицу шагов 14–19 вверху и раздел «Этап 1 — статус»).

### На каком этапе мы сейчас

- По `description.md` завершённые работы попадают в **Этап 1**.
- Переход к **Этапу 2** — авторизация и роли (см. `description.md`), после фиксации текущего кода коммитом.

### Алгоритм работы над этапами

Подтверждено принятие для дальнейших итераций: **План → жду «ОК» → код → тесты (unit + интеграционные) → golangci-lint / ESLint → отчёт (структура, команды, текст коммита).**

---

## Предложенная структура репозитория (целевая для конца Этапа 1)

```
.
├── cmd/server/              # точка входа API
├── internal/
│   ├── app/
│   ├── config/
│   ├── db/
│   ├── models/
│   ├── repository/
│   └── router/
├── migrations/              # 001_* .up.sql / .down.sql
├── frontend/                # статика TMA + npm только для ESLint/Jest позже
│   ├── package.json
│   ├── eslint.config.*
│   ├── index.html
│   ├── src/ или корневые app.js, styles.css
│   └── (позже этапы 4–5: тесты)
├── .github/workflows/       # ci.yml — Go тесты, lint, опционально job для frontend lint
├── description.md           # ТЗ и этапы
├── process.md               # текущий прогресс
├── go.mod
├── .golangci.yml
├── .gitignore
├── docker-compose.yml       # минимально: Postgres для локальной разработки и интеграционных тестов
└── README.md                # как поднять БД и запускать сервер
```

На **Этапе 7** логично расширить: `Dockerfile` для API, nginx для статики и полноценный compose для прод-сценария; на Этапе 1 достаточно compose с Postgres (и при желании уже сервис `api`), чтобы не блокировать интеграционные тесты.

---

## Команды инициализации (справочно: часть уже сделана в этом репо)

Уже есть: `git init`, `go mod` с модулем `github.com/example/coworking`.

Для привязки к GitHub (пример):

```powershell
git remote add origin https://github.com/<org>/<repo>.git
git branch -M main
git push -u origin main
```

Для frontend-линтинга (выполнить после появления `frontend/`):

```powershell
cd frontend
npm init -y
npm install -D eslint @eslint/js globals
# при выборе Jest для этапов 4–5: npm install -D jest --save-dev
```

---

## На согласование: подробный план «следующий шаг» — закрытие пунктов Этапа 1

Дальнейшая работа **после вашего «ОК»** по этому плану:

1. **Миграции PostgreSQL**
   - Файлы: `migrations/001_initial.up.sql`, `migrations/001_initial.down.sql`.
   - Таблицы в соответствии с описанием: `rooms`, `workspaces`, `users`, `bookings`, `booking_settings`; связи FK, ограничения CHECK ролей/статусов (минимальный набор, расширяемый на Этапах 2–3).
   - Индексы: уникальность логических ключей там, где нужно (например `telegram_id` для студента уже в модели — отразить в SQL).

2. **Применение миграций в dev и CI**
   - Вариант А: утилита `golang-migrate/migrate` в `go.mod` + `Makefile` или скрипт `go run ./cmd/migrate` с DSN из env.
   - Вариант Б (минимализм Этапа 1): только SQL-файлы + документированный шаг «применить через psql/docker» + в CI выполнить те же файлы перед `go test`.
   - Рекомендация в плане: один явный способ (предпочтительно А), чтобы тест «подключение + схема» был воспроизводимым.

3. **Интеграционные тесты БД**
   - Пакет, например `internal/db/integration_test.go` с build-tag `integration` или отдельный каталог `test/integration` + `docker compose`-сервис в CI через `services: postgres`.
   - Тесты: успешное `Ping`/подключение, транзакция create/drop через тестовую таблицу или наличие таблицы после migrate.
   - В `go test` по умолчанию — быстрые unit-тесты; полный интеграционный прогон: `go test -tags=integration ./...` (точную команду зафиксировать в README и CI).

4. **GitHub Actions**
   - Файл: `.github/workflows/ci.yml`: checkout, Setup Go `1.25`, кэш модулей, сервис PostgreSQL, шаг применения миграций, `golangci-lint run`, `go test ./...` (+ опционально job с ESLint когда появится `frontend/package.json`).
   - Секреты не обязательны на Этапе 1.

5. **Локальный Postgres**
   - `docker-compose.yml`: сервис `db` (образ Postgres 15+), том, порты; переменные `POSTGRES_*` согласованы с `DB_DSN` в `internal/config`.

6. **Frontend-скелет TMA и ESLint**
   - Файлы: `frontend/index.html`, подключение `https://telegram.org/js/telegram-web-app.js`, `Telegram.WebApp.ready()`, расширение на всю высоту, продовые цвета/типографика без отладочного текста.
   - `package.json` + `eslint.config.js` (flat config, ESLint 9), игнор `telegram-web-app` глобалов через `globals`.
   - Отдельно: решить, нужен ли **Prettier** на Этапе 1 (по описанию — да; добавить конфиг без лишних правил).

7. **Документация**
   - Обновление `README.md`: как поднять БД (`docker compose up -d`), DSN для локального запуска, команды тестов и линтеров (`golangci-lint run ./...`, `npm run lint` в `frontend`).

### Команды качества после реализации (для финального отчёта этапа)

Backend:

```powershell
cd c:\Users\mitri\Desktop\ПрогИнж
go mod tidy
go test ./...
golangci-lint run ./...
```

Frontend (после `npm install`):

```powershell
cd frontend
npm run lint
```

*(Точное имя npm-скрипта задаётся в `package.json`.)*

---

*Дальнейшие записи — после утверждения плана и реализации кода.*

---

*Отчёт будет обновляться по мере завершения каждого подпункта.*

---

## План Этапа 5 (Frontend: Админ-панель и Схемы)

1. **Новые файлы Frontend**:
   - `frontend/admin.html`: Отдельная точка входа для администратора (полноценный дашборд, без TMA-обвязки, так как админ заходит с ПК/планшета по email).
   - `frontend/admin.js`: Логика авторизации, роутинга вкладок и взаимодействия с API.
   - `frontend/admin.css`: Стили админ-панели (переиспользование UI-компонентов, но адаптация под широкий экран).
   
2. **Интерфейс Авторизации**:
   - Форма входа по Email/Паролю.
   - Получение и сохранение JWT в `localStorage`.
   
3. **Дашборд Управления**:
   - Вкладка **"Помещения"**: Список комнат, добавление/редактирование/удаление.
   - Вкладка **"Схемы (Рабочие места)"**: Выбор помещения, интерфейс добавления рабочих мест. Будем использовать Canvas или простой SVG/Grid интерфейс с drag & drop или кликом для добавления стола/места.
   - Вкладка **"Бронирования"**: Просмотр всех бронирований в системе, возможность отмены администратором.
   - Вкладка **"Отчеты (Статистика)"**: Агрегация данных по загруженности коворкинга на основе загрузки бронирований и генерация простейшего графика/таблицы (расчеты на фронтенде).

4. **Тесты и Линтинг**:
   - Написание тестов (`admin.test.js`) для логики трансформации данных статистики/отчетов.
   - Прогон кода через ESLint (`npm run lint`).
