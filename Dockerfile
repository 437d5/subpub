FROM golang:1.24.3-alpine3.21 AS builder

WORKDIR /app
COPY . .

RUN go mod download && go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/main/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server .

ENV GRPC_PORT=5000

EXPOSE ${GRPC_PORT}

CMD [ "./server" ]