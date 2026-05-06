## 📋 Описание проекта

REST API для банковского сервиса на языке Go с поддержкой:
- Аутентификации и авторизации (JWT)
- Управления банковскими счетами
- Переводов между счетами
- Выпуска банковских карт (с шифрованием)
- Кредитования с аннуитетными платежами
- Автоматического шедулера для просрочек
- Аналитики и прогнозирования баланса
- Интеграции с ЦБ РФ (SOAP)
- Email уведомлений (SMTP)

### Требования

- Go 1.21 или выше
- PostgreSQL 17 или выше

### Установка и запуск

1. **Клонировать репозиторий**
```bash
git clone https://github.com/YOUR_USERNAME/bank-api.git
cd bank-api
Настроить переменные окружения

bash
cp .env.example .env
# Отредактируйте .env под ваши настройки
Создать базу данных

sql
CREATE DATABASE bank_app;
Выполнить миграцию

bash
psql -U postgres -d bank_app -f migrations/001_full_init.sql
Установить зависимости

bash
go mod tidy
Запустить сервер

bash
go run cmd/main.go
Сервер запустится на http://localhost:8080

📡 API Endpoints
Публичные (без JWT)
Метод	Endpoint	Описание
POST	/api/register	Регистрация пользователя
POST	/api/login	Вход (возвращает JWT токен)
GET	/health	Проверка работы сервера
Защищенные (требуют JWT)
Метод	Endpoint	Описание
POST	/api/accounts	Создать счет
GET	/api/accounts	Получить все счета
POST	/api/deposit	Пополнить счет
POST	/api/withdraw	Снять деньги
POST	/api/transfer	Перевести между счетами
GET	/api/transactions/{accountId}	История транзакций
POST	/api/cards	Выпустить карту
GET	/api/cards?account_id={id}	Получить карты счета
POST	/api/credits/apply	Оформить кредит
GET	/api/credits/{creditId}/schedule	График платежей
GET	/api/analytics/dashboard	Дашборд аналитики
GET	/api/analytics/credit-burden	Кредитная нагрузка
GET	/api/accounts/{accountId}/predict?days=30	Прогноз баланса
GET	/api/central-bank/rate	Ключевая ставка ЦБ РФ
POST	/api/email/test	Тест email уведомлений

🧪 Тестирование API
Способ 1: Через файл test-api.http (VS Code)
Установите расширение REST Client в VS Code

Откройте файл test-api.http

Выполняйте запросы последовательно, нажимая Send Request

Способ 2: Через curl
bash
# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","username":"test","password":"123456"}'

# Логин (сохраните token)
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"123456"}'

# Создать счет
curl -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"currency":"RUB"}'

# Пополнить счет
curl -X POST http://localhost:8080/api/deposit \
  -H "Authorization: Bearer <YOUR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"account_id":"<ACCOUNT_ID>","amount":10000,"description":"Пополнение"}'
Способ 3: Через Postman
Импортируйте коллекцию из файла test-api.http или создайте запросы вручную.

📊 Примеры ответов
Успешный логин (200 OK)
json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "test@test.com",
    "username": "test",
    "created_at": "2026-05-06T20:00:00Z"
  }
}
Создание счета (201 Created)
json
{
  "message": "Account created successfully",
  "account": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "balance": 0,
    "currency": "RUB",
    "created_at": "2026-05-06T20:00:01Z"
  }
}
Оформление кредита (201 Created)
json
{
  "message": "Credit approved!",
  "credit": {
    "id": "770e8400-e29b-41d4-a716-446655440002",
    "amount": 50000,
    "interest_rate": 15,
    "monthly_payment": 9025.83,
    "status": "active",
    "created_at": "2026-05-06T20:00:02Z"
  }
}
Ключевая ставка ЦБ РФ (200 OK)
json
{
  "rate": 16.0,
  "bank_margin": 5.0,
  "final_rate": 21.0,
  "updated_at": "2026-05-06 20:00:03"
}

🏗 Архитектура проекта
text
bank-api/
├── cmd/main.go                 # Точка входа
├── internal/
│   ├── models/                 # Модели данных
│   ├── repositories/           # Работа с БД
│   ├── services/               # Бизнес-логика
│   ├── handlers/               # HTTP обработчики
│   ├── middleware/             # JWT проверка
│   ├── scheduler/              # Шедулер (просрочки)
│   └── utils/                  # Утилиты (JWT, шифрование)
├── pkg/database/               # Подключение к БД
├── migrations/                 # SQL миграции
├── .env.example                # Шаблон конфигурации
└── test-api.http               # Тесты API

🔒 Безопасность
JWT токены (24 часа)

bcrypt для хеширования паролей

HMAC-SHA256 для целостности карт

PGP шифрование для данных карт (опционально)

Middleware для защиты маршрутов

⚙️ Настройка переменных окружения (.env)
env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=bank_app

# JWT Secret
JWT_SECRET=super-secret-jwt-key-2024

# HMAC Secret
HMAC_SECRET=hmac-super-secret-key-2024

# SMTP (для email уведомлений)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password

🛠 Технологии
Компонент	Технология
Язык	Go 1.21
БД	PostgreSQL 17
Маршрутизация	gorilla/mux
Аутентификация	JWT (golang-jwt/jwt)
Хеширование	bcrypt, HMAC-SHA256
Шифрование	PGP (ProtonMail/go-crypto)
XML парсер	beevik/etree
Email	go-mail/mail/v2
Логирование	logrus

👤 Автор
Dmitry Zaytsev
