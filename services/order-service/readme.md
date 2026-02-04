go install github.com/pressly/goose/v3/cmd/goose@latest

 создание миграции 
 go install github.com/pressly/goose/v3/cmd/goose@latest


goose -dir ./migrations create create_orders_table sql

goose -dir ./migrations postgres "postgres://postgres:password@localhost:5434/orders?sslmode=disable" up