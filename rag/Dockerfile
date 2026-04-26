FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o rag-server ./cmd/server/main.go

FROM golang:1.24 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /rag-server ./cmd/server/main.go

FROM python:3.10-slim
WORKDIR /app
RUN apt-get update && apt-get install -y tesseract-ocr tesseract-ocr-rus libgl1 libglib2.0-0 poppler-utils && rm -rf /var/lib/apt/lists/*
RUN pip3 install --no-cache-dir pytesseract pdf2image numpy opencv-python-headless pdfplumber
RUN pip3 install --no-cache-dir torch torchvision --index-url https://download.pytorch.org/whl/cpu
RUN pip3 install --no-cache-dir layoutparser iopath pyyaml
RUN pip3 install --no-cache-dir playwright && playwright install chromium --with-deps
COPY --from=builder /rag-server /rag-server
COPY ./migrations ./migrations
COPY ./python_scripts ./python_scripts
EXPOSE 50051
ENTRYPOINT ["/rag-server"]