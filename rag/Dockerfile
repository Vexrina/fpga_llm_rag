FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o rag-server ./cmd/server/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/rag-server /app/rag-server
COPY ./migrations ./migrations
EXPOSE 50051
ENTRYPOINT ["/app/rag-server"] 