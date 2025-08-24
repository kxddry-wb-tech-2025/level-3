# use bookworm for delve
FROM golang:1.25.0-bookworm AS builder

WORKDIR /app
ENV CGO_ENABLED=0


COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY src/ ./src/

# debugging
# RUN go build -gcflags "all=-N -l" -o app ./src/cmd/booker/

# production
RUN go build -o app ./src/cmd/booker/

FROM golang:1.25.0
WORKDIR /app

# RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=builder /app/app .

# debug
# EXPOSE 40000
# CMD ["dlv", "exec", "./app", "--headless", "--listen=:40000", "--api-version=2", "--accept-multiclient", "--log"]

CMD ["./app"]