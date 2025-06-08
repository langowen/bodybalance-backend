# BodyBalance Backend API

![Coverage](https://img.shields.io/badge/Coverage-35.1%25-green)
![Go Version](https://img.shields.io/badge/Go-1.24.3-blue)
![Latest Release](https://img.shields.io/badge/Release-0.3.1-blue)

Бэкенд-сервис для приложения по медицинской реабилитации. API обеспечивает доступ к видео-контенту, разделенному по категориям.
Репозиторий Android приложения: [BodyBalance](https://github.com/DecardCain21/BodyBalance)
## Функции

- Каталог видео, категорий, типов контента
- Аутентификация пользователей
- Административный интерфейс для управления контентом
- Документация API (Swagger)

## Зависимости

- Go 1.24.3
- PostgresSQL 16
- Redis

## Интерфейс администратора
Административный интерфейс доступен по адресу `/admin/web` после запуска сервиса, данные для входа: `DOCS_USER:DOCS_PASSWORD`.
Из административного интерфейса можно управлять видео-контентом, категориями, типами контента и пользователями.

## Документация

Документация API доступна по адресу `/admin/docs` после запуска сервиса, данные для входа: `DOCS_USER:DOCS_PASSWORD`.

## Запуск

### Локально

```bash
# Настройка конфигурации
cp env.example .env
# Редактирование конфигурации
nano .env
# Редактирование конфигурации
nano config/local.yaml
# Запуск сервиса
go run ./cmd/bodybalance
```

### Docker



```bash
# Редактирование volumes и конфигурации
nano docker-compose.yml
# Настройка конфигурации
cp env.example .env
# Редактирование конфигурации
nano .env
# Запуск с помощью Docker Compose
docker-compose up -d
```
