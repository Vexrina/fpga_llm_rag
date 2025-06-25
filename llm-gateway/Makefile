deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "Dependencies installed"


proto: deps
	mkdir -p pkg/llm-gateway
	protoc --go_out=pkg --go-grpc_out=pkg \
		api/gateway.proto
	@echo "Proto files generated"

# Генерация swagger/openapi документации
swagger: proto
	protoc -Iapi \
       --openapiv2_out=pkg/llm-gateway \
       --openapiv2_opt logtostderr=true \
       api/gateway.proto

clean:
	rm -rf bin/
	rm -rf pkg/llm-gateway/*.pb.go
	rm -rf pkg/llm-gateway/*.swagger.json
	@echo "Clean complete"

generate: clean deps proto swagger
	@echo "Full generation completed"

FW_PROTO_DIR=../float-weaver/api
FW_PROTO_FILE=floatweaver.proto
FW_OUT_DIR=./pkg/

gen-floatweaver:
	mkdir -p $(FW_OUT_DIR)
	protoc \
	  --proto_path=$(FW_PROTO_DIR) \
	  --go_out=$(FW_OUT_DIR) \
	  --go-grpc_out=$(FW_OUT_DIR) \
	  $(FW_PROTO_DIR)/$(FW_PROTO_FILE)

RAG_PROTO_DIR=../rag/api
RAG_PROTO_FILE=rag.proto
RAG_OUT_DIR=./pkg/

gen-rag:
	mkdir -p $(RAG_OUT_DIR)
	protoc \
	  --proto_path=$(RAG_PROTO_DIR) \
	  --go_out=$(RAG_OUT_DIR) \
	  --go-grpc_out=$(RAG_OUT_DIR) \
	  $(RAG_PROTO_DIR)/$(RAG_PROTO_FILE)