# float-weaver

GRPC сервис для эмбеддинга текста с помощью модели mxbai-embed-large на Go.

## Быстрый старт

1. Скачайте модель mxbai-embed-large с HuggingFace:
   https://huggingface.co/mxbai-embed-large
   и поместите файлы в папку `models/mxbai-embed-large` (должны быть файлы config.json, model.safetensors и т.д.).
2. Запустите сервис TGI:
   ```sh
   docker compose up -d
   ```
3. TGI будет доступен на http://localhost:8080

## Основные компоненты
- `api/floatweaver.proto` — описание gRPC API
- `cmd/server/main.go` — запуск сервера
- `internal/` — бизнес-логика
- `generated/` — сгенерированные protobuf-файлы

## Как работает
1. Получает текст через gRPC.
2. Отправляет его в модель mxbai-embed-large (локально или через HTTP API).
3. Возвращает эмбеддинг (массив float32). 