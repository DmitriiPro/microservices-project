.PHONY: all user

PROTO_SRC=.
GOOGLEAPIS_DIR=./google
VALIDATE_DIR=./validate
PROTO_OUT=../services/user-service/internal/pb
SWAGGER_OUT=../services/user-service/swagger
PROTOC=protoc

all: user

user:
	@echo "Generating Go + gRPC + Gateway + Validator..."
	mkdir -p $(PROTO_OUT)
	mkdir -p $(SWAGGER_OUT)

	$(PROTOC) \
		--proto_path=$(PROTO_SRC) \
		--proto_path=$(GOOGLEAPIS_DIR) \
		--proto_path=$(VALIDATE_DIR) \
		--go_out=paths=source_relative:$(PROTO_OUT) \
		--go-grpc_out=paths=source_relative:$(PROTO_OUT) \
		--grpc-gateway_out=paths=source_relative:$(PROTO_OUT) \
		--validate_out=lang=go,paths=source_relative:$(PROTO_OUT) \
		--openapiv2_out=$(SWAGGER_OUT) \
		--openapiv2_opt logtostderr=true \
		user/user.proto


	@echo "Done!"
