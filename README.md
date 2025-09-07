# Auth Service

Универсальный сервис аутентификации и авторизации для ShopFlow.  
Обеспечивает вход и выдачу JWT токенов для всех микросервисов, таких как **Application**, **Notification**, **Order**, **Gateway** и др.  
Все микросервисы используют этот сервис для проверки токенов и идентификации пользователей.

---

## Стек

- Go 1.21+
- Gin
- JWT (github.com/golang-jwt/jwt/v5)
- PostgreSQL (GORM)
- RabbitMQ (опционально для событий)

---

## Установка

1. Клонировать репозиторий:
```bash
git clone <URL_REPO>
cd auth



