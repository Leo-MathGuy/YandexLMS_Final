# Yandex Go Course Final Project

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Leo-MathGuy/YandexLMS_Final/go.yml?label=tests)](https://github.com/Leo-MathGuy/YandexLMS_Final/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/Leo-MathGuy/YandexLMS_Final/badge.svg?branch=main)](https://coveralls.io/github/Leo-MathGuy/YandexLMS_Final?branch=main)

[На русском](README_ru.md)

This is the final project that took two weeks of great effort to complete.
This is a demonstration of web development, REST API, JWT authentication, SQL database usage, and GRPC communication in Go,
written from scratch in the form of a calculator app.

## Running

A main.go from the project root directory launches both the app and the agent. you can run it with

```bash
go run main.go
```

Alternatively, you can run the two programs separately (not recommended, as the helper script initializes shared logs):

```bash
# Terminal 1
go run cmd/app/main.go

# Terminal 2
go run cmd/agent/main.go
```

## Testing

### Tests

Unit, integration and race condition tests can be run for the entire project by running the following command:

WARNING: Running tests will delete the database

```bash
go test ./... -race
```

The frontend is styled and has a user friendly interface. The entire program can be tested that way. You can reach it on [:8080](localhost:8080)

### Testing the API

You can test the API with CURL. Commands are in the [API Documentation](#api-endpoints)

## Documentation

### Application (app)

Purpose:

* Hosts frontend
* Hosts REST API
* Processes expressions into an AST for later calculation by the agent
* Manages and serves tasks
* Saves to and loads from a sqlite database
* Serves a gRPC server for the agent

Features:

* Logging
* In-depth error handling
* Unit and integration tests

### Agent (agent)

Purpose:

* Multithreaded calculation of individual tasks
* Recieves tasks from the app

Features:

* Logging
* Tests are done by the server intergration tests

#### API Endpoints

##### POST /api/v1/register

Register user.

Rules:

* Username not case-sensitive
* Username must start with letter, but can contain numbers
* Username and password must be 3-32 length

Test command:

```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "bob",
  "password": "123"
}'
```

Error:

```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "123",
  "password": "123"
}'
```

##### POST /api/v1/login

Log into account. Returns a JWT token as the response body and a set-cookie for browsers

Test command:

```bash
curl --location 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
  "login": "bob",
  "password": "123"
}'
```

Error:

```bash
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
  "login": "eve",
  "password": "123"
}'
```

##### POST /api/v1/calculate

Send expression. Must be logged in. Takes in token in json body

Test command:

```bash
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2",
  "token": "
}'
```

Returns:

```json
{"id":<id>}
```

##### GET /api/v1/expressions

Returns expressions belonding to the currently logged in user. Takes in token as Authentication header

Test command:

```bash
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authentication: <token>' \
```

Returns:

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

Returns expression under the given id, if it belongs to the currently logged in user. Takes in token as Authentication header

Test command:

```bash
curl --location 'localhost:8080/api/v1/expressions/1' \
--header 'Authentication: <token>' \
```

Returns:

```json
{"expression":{"id":1,"result":59,"status":true}} // Ready
```

```json
{"expression":{"id":2,"result":0,"status":false}} // Processing
```

##### /sans

sans.
