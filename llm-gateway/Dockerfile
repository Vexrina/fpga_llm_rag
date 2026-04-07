FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf protoc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/server /bin/server

EXPOSE 8083

ENTRYPOINT ["/bin/server"]