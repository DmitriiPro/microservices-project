.PHONY: all user

PROTO_SRC=.
GOOGLEAPIS_DIR=./googleapis
PROTO_OUT=../services/user-service/internal/pb
PROTOC=protoc

all: user

user:
	@echo "Generating Go + gRPC + Gateway..."
	mkdir -p $(PROTO_OUT)

	$(PROTOC) \
		--proto_path=$(PROTO_SRC) \
		--proto_path=$(GOOGLEAPIS_DIR) \
		--go_out=$(PROTO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_OUT) \
		--grpc-gateway_opt=paths=source_relative \
		user/user.proto

	@echo "Done!"
