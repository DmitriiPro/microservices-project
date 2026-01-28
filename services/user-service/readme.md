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

//создание файла для миграции 
migrate create -ext sql -dir ./migrations -seq create_users_table

Поднимаем контейнеры:

docker-compose up -d


Применяем миграции:

migrate -path ./migrations -database "postgres://user:password@localhost:5432/users" up


Запускаем сервис:

go run cmd/main.go

docker exec -it user_postgres psql -U user -d users
