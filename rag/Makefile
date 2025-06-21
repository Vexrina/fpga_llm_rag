.PHONY: proto build run clean swagger

# Установка зависимостей
deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go get -u github.com/go-jet/jet/v2
	go install github.com/go-jet/jet/v2/cmd/jet@latest
	export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Dependencies installed"

# Генерация Go кода из proto файлов
proto: deps
	mkdir -p pkg/rag
	protoc --go_out=pkg --go-grpc_out=pkg \
		api/rag.proto
	@echo "Proto files generated"

# Генерация swagger/openapi документации
swagger: proto
	protoc -Iapi \
       --openapiv2_out=pkg/rag \
       --openapiv2_opt logtostderr=true \
       api/rag.proto

# Сборка проекта
build: proto
	go build -o bin/server cmd/server/main.go
	@echo "Project built"

# Запуск сервера
run: build
	./bin/server

# Очистка
clean:
	rm -rf bin/
	rm -rf pkg/rag/*.pb.go
	rm -rf pkg/rag/*.swagger.json
	@echo "Clean complete"

# Полная генерация (все зависимости + очистка + генерация)
generate: clean deps proto swagger
	@echo "Full generation completed"

# Генерация схем используя jet
jet: deps
	jet -source=postgresql -host=localhost -port=5432 -user=rag_user -password=rag_password -dbname=rag_db -schema=public -path=generated/
