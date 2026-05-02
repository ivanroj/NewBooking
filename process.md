# Процесс разработки – Этап 1 (Bootstrap)

## Что уже выполнено

| Шаг | Описание | Файл/Компонент |
|-----|----------|----------------|
| 1️⃣  | Инициализация репозитория Git | `git init` (создан `.git` каталог) |
| 2️⃣  | Базовый README с обзором проекта | `README.md` |
| 3️⃣  | .gitignore с типовыми записями для Go, Node, IDE, OS | `.gitignore` |
| 4️⃣  | Go‑модуль и зависимости (`sqlx`, `gorilla/mux`, `jwt-go`) | `go.mod` |
| 5️⃣  | Перемещение точки входа в `cmd/server/main.go` (корневой `main.go` оставлен как заглушка) | `cmd/server/main.go`, `main.go` |
| 6️⃣  | Загрузчик конфигурации (порт, DSN) | `internal/config/config.go` |
| 7️⃣  | Базовый bootstrap‑слой приложения (создание DB, репозитория, роутера) | `internal/app/app.go` |
| 8️⃣  | Обёртка подключения к PostgreSQL через `sqlx` | `internal/db/postgres.go` |
| 9️⃣  | Модели данных: `Room`, `Workspace`, `User`, `Booking`, `BookingSettings` | `internal/models/models.go` |
| 🔟  | Репозиторий с заглушкой получения всех помещений | `internal/repository/repo.go` |
| 1️⃣1️⃣ | Роутер с эндпоинтом `/health` | `internal/router/router.go` |
| 1️⃣2️⃣ | Конфигурация golangci‑lint (включены основные линтеры) | `.golangci.yml` |
| 1️⃣3️⃣ | Добавлены скрипты‑команды для дальнейшего создания миграций, Docker‑compose и CI‑pipeline (см. план) | — |

## Что планируется к следующему шагу (в рамках Этапа 1)

- **Миграции**: создать `migrations/001_init.up.sql` и `down.sql` с таблицами `rooms`, `workspaces`, `users`, `bookings`, `booking_settings`.
- **Docker‑compose**: файл `docker-compose.yml` с сервисами `db` (Postgres) и `api` (Go‑бинарник).
- **CI/CD**: GitHub Actions workflow `ci.yml` – запуск тестов и линтеров.
- **Тесты**: минимальный unit‑тест здоровья сервера (`/health`).
- **Frontend‑скелет**: `frontend/index.html`, `app.js`, `styles.css` – базовый UI‑шаблон Telegram Mini App.
- **Запуск**: `docker compose up -d` и проверка `curl http://localhost:8080/health`.

## Команды для быстрой реализации оставшихся пунктов (скопировать и выполнить)
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

---

*Отчёт будет обновляться по мере завершения каждого подпункта.*
