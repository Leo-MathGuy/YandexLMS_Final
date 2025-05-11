# Финальный проект курса Yandex Go

[![Статус рабочего процесса GitHub Actions](https://img.shields.io/github/actions/workflow/status/Leo-MathGuy/YandexLMS_Final/go.yml?label=tests)](https://github.com/Leo-MathGuy/YandexLMS_Final/actions/workflows/go.yml)
[![Статус покрытия](https://coveralls.io/repos/github/Leo-MathGuy/YandexLMS_Final/badge.svg?branch=main)](https://coveralls.io/github/Leo-MathGuy/YandexLMS_Final?branch=main)

Это итоговый проект, над которым я работал 2 недели
.
Проект демонстрирует веб-разработку, REST API, аутентификацию по JWT, работу с SQL-базой данных и взаимодействие по gRPC на языке Go, написан с нуля в формате калькулятор-приложения.

## Запуск

Файл `main.go` в корневой директории проекта запускает одновременно и приложение, и агент. Запустить его можно командой:

```bash
go run main.go
```

Альтернативно можно запустить обе программы отдельно (не рекомендуется, так как вспомогательный скрипт main.go инициализирует общие логи):

```bash
# Терминал 1
go run cmd/app/main.go

# Терминал 2
go run cmd/agent/main.go
```

## Тестирование

### Тесты

Модульные и интеграционные тесты с проверкой гонок можно запустить для всего проекта командой:

**ВНИМАНИЕ:** запуск тестов приведёт к удалению базы данных

```bash
go test ./... -race
```

Фронтенд стилизован и имеет удобный интерфейс, что позволяет тестировать программу в браузере.

### Тестирование API

API можно протестировать с помощью CURL. Команды приведены в документации по API.

## Документация

### Приложение (app)

Назначение:

* Хостинг фронтенда
* Хостинг REST API
* Преобразование выражений в AST для последующих вычислений агентом
* Управление и выдача задач
* Сохранение и загрузка данных из SQLite базы данных
* Запуск gRPC-сервера для агента

Особенности:

* Логирование
* Глубокая обработка ошибок
* Модульные и интеграционные тесты

### Агент (agent)

Назначение:

* Многопоточное вычисление отдельных задач
* Получение задач от приложения

Особенности:

* Логирование
* Тестирование проводится в интеграционнах тестах сервера

#### API

##### POST /api/v1/register

Зарегистрируйте пользователя.

Правила:

* Имя пользователя не чувствительно к регистру
* Имя пользователя должно начинаться с буквы, но может содержать цифры
* Имя пользователя и пароль должны быть длиной от 3 до 32 символов

Тестовая команда:

```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "bob",
  "password": "123"
}'
```

##### POST /api/v1/login

Войти в учетную запись. Возвращает токен JWT в качестве тела ответа и set-cookie для браузеров

Тестовая команда:

```bash
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "bob",
  "password": "123"
}'
```

##### POST /api/v1/calculate

Отправить выражение. Необходимо войти в систему. Принимает токен в теле json

Тестовая команда:

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2",
  "token": "
}'
```

Возврат:

```json
{"id":<id>}
```

##### GET /api/v1/expressions

Возвращает выражения, принадлежащие текущему вошедшему в систему пользователю. Принимает токен как заголовок аутентификации

Тестовая команда:

```bash
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authentication: <token>' \
```

Возврат:

```json
{
    "expressions":[
        {"id":1,"result":59,"status":true}, // Ready
        {"id":2,"result":0,"status":false} // Processing
        // And so on
    ]
}
```

##### GET /api/v1/expressions/{id}

Возвращает выражение под указанным идентификатором, если оно принадлежит текущему вошедшему в систему пользователю. Принимает токен как заголовок аутентификации

Тестовая команда:

```bash
curl --location 'localhost:8080/api/v1/expressions/1' \
--header 'Authentication: <token>' \
```

Возврат:

```json
{"expression":{"id":1,"result":59,"status":true}} // Ready
```

```json
{"expression":{"id":2,"result":0,"status":false}} // Processing
```
