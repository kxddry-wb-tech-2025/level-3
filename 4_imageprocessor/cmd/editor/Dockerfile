FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o editor ./cmd/editor/main.go

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/editor /app/editor

CMD ["/app/editor"]