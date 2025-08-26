FROM golang:1.25.0 AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY src/ ./src/

RUN go build -o app ./src/cmd/app/

FROM golang:1.25.0
WORKDIR /app

COPY --from=builder /app/app .

CMD ["./app"]