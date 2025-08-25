FROM golang:1.25.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -o app ./cmd/notifier/main.go

FROM golang:1.25.0
WORKDIR /app

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY --from=builder /app/app /app/app
COPY static/ ./static/

EXPOSE 40000

CMD ["dlv", "--listen=:40000", "--headless=true", "--api-version=2", "exec", "/app/app"]