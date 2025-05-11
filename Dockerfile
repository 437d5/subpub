FROM golang:1.24.3-alpine3.21 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -0 App ./cmd/main/main.go

FROM alpine:3.21

COPY --from=builder App .

EXPOSE 5000

CMD [ "App" ]