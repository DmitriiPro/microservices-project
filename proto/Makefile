.PHONY: all user

# Путь к исходным proto
PROTO_SRC=.

# Куда генерируем Go-код (внутри user-service)
PROTO_OUT=../services/user-service/internal/pb

# Префикс для пакета
PROTO_PKG=userv1

# Протокол компилятор (если в PATH)
PROTOC=protoc

# Все proto
PROTO_FILES=$(shell find $(PROTO_SRC)/user -name "*.proto")

# По умолчанию генерируем user
all: user

user:
	@echo "Generating Go files from proto..."
	mkdir -p $(PROTO_OUT)
	protoc \
		--proto_path=$(PROTO_SRC) \
		--go_out=$(PROTO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) \
		--grpc-gateway_opt=paths=source_relative \
		user/user.proto
	@echo "Done!"
