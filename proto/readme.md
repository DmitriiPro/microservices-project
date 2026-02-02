Устанавливаем инструменты

Нам нужны:

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.3

go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.3

# для валидации в файле proto
go install github.com/envoyproxy/protoc-gen-validate@latest

Проверка: 

protoc-gen-go --version
protoc-gen-go-grpc --version
protoc-gen-grpc-gateway --version
protoc-gen-openapiv2 --version


Использование Makefile

Открой Git Bash или PowerShell:

cd proto
make


Нужно скачать google api protos (1 раз)

В папке proto/:

git clone https://github.com/googleapis/googleapis

Скачать .proto файлы PGV

PGV хранит свои .proto в репозитории:

# Перейдем в proto-директорию проекта
cd proto/

# Скачиваем PGV proto файлы
git clone https://github.com/envoyproxy/protoc-gen-validate.git validate

go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
