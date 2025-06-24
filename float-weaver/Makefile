# Установка зависимостей
deps:
	go mod tidy
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	# go get -u github.com/go-jet/jet/v2
	# go install github.com/go-jet/jet/v2/cmd/jet@latest
	export PATH=$PATH:$(go env GOPATH)/bin
	@echo "Dependencies installed"

# Генерация Go кода из proto файлов
proto: deps
	mkdir -p pkg/floatweaver
	protoc --go_out=pkg --go-grpc_out=pkg \
		api/floatweaver.proto
	@echo "Proto files generated"

swagger: proto
	protoc -Iapi \
       --openapiv2_out=pkg/floatweaver \
       --openapiv2_opt logtostderr=true \
       api/floatweaver.proto

clean:
	rm -rf bin/
	rm -rf pkg/floatweaver/*.pb.go
	rm -rf pkg/floatweaver/*.swagger.json
	@echo "Clean complete"

generate: clean deps proto swagger
	@echo "Full generation completed"

.PHONY: download-model
download-model:
	mkdir -p models/mxbai-embed-large
	huggingface-cli download mixedbread-ai/mxbai-embed-large-v1 --local-dir models/mxbai-embed-large-v1 --local-dir-use-symlinks False
