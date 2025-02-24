# Сервис аутентификации (Auth Service)

Этот репозиторий содержит микросервис для регистрации пользователей, авторизации и валидации токенов. Проект разработан на Go с использованием gRPC и PostgreSQL и используется в качестве вспомогательного сервиса в другом проекте [Team Manager](https://github.com/hard-gainer/team-manager).

## Возможности
1. Регистрация пользователей (с возможностью назначать роль администратора).  
2. Авторизация на основе JWT.  
3. Проверка прав (IsAdmin) и валидация токенов.  
4. Хранение пользователей и приложений в базе PostgreSQL.

## Технологии
- Go + gRPC  
- PostgreSQL (через драйвер pgx)  
- Docker для контейнеризации  
- Миграции с помощью [golang-migrate](https://github.com/golang-migrate/migrate)

## Взаимодействие с другим сервисом
- Этот сервис предоставляет RPC-методы, которые вызываются другим микросервисом.  
- Запросы и ответы основаны на gRPC-протоколе, позволяя быстро и безопасно проверять или выдавать JWT-токены.

## Запуск
1. Клонируйте репозиторий.  
2. Укажите настройки в файле config.yaml (URL к базе данных и порт gRPC).  
3. Примените миграции:
    ```
        make createdb
        make migrateup
    ```