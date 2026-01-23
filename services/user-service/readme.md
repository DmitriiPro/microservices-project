user-service/
├── cmd/
│   └── main.go          # точка входа
├── internal/
│   ├── handler/         # gRPC handlers
│   ├── service/         # бизнес-логика
│   ├── repository/      # Postgres
│   ├── model/           # доменные модели
│   └── config/
├── migrations/          # SQL миграции
└── go.mod
