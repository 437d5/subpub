# PubSub gRPC микросервис
Приложение проверено в Postman и написаны unit-тесты для пакета subpub. Реализован Graceful Shutdown.

- Конфигурация через переменные окружения
- Оптимизированный Docker образ

## Инструкция по запуску 

 1. Склонировать этот репозиторий 
```
git clone https://github.com/437d5/subpub.git
```
2. Собрать докер образ 
```
docker build -t pubsub .
```
3. Запустить контейнер с приложением 
```
docker run -p 5000:5000 pubsub
```

## Структура проекта
```
├── api 
│   └── proto
│       └── subpub.proto       # Здесь лежит proto файл сервиса
├── cmd
│   └── main
│       └── main.go
├── Dockerfile 
├── go.mod
├── go.sum
├── internal
│   ├── app
│   │   └── app.go              # Объявление и методы для работы с App
│   ├── config
│   │   └── config.go           # Работа с переменными окружения
│   └── server
│       └── server.go           # Реализация интерфейса сервера gRPC
└── pkg
    ├── pb                      # Сгенерированные файлы
    │   ├── subpub_grpc.pb.go
    │   └── subpub.pb.go
    └── subpub                  # Первая часть задания. Реализация пакета subpub
        ├── errors.go           # Объявления кастомных ошибок
        ├── subpub.go 
        └── subpub_test.go      # Тесты
```