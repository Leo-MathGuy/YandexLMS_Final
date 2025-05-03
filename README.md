# Yandex Go Course Final Project

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Leo-MathGuy/YandexLMS_Final/go.yml?label=tests)](https://github.com/Leo-MathGuy/YandexLMS_Final/actions/workflows/go.yml)
![Coveralls](https://img.shields.io/coverallsCoverage/github/Leo-MathGuy/YandexLMS_Final)


This is the final project that took two weeks of great effort to complete.
This is a demonstration of web development, REST API, JWT authentication, SQL database usage, and GRPC communication in Go,
written from scratch in the form of a calculator app.

## Running

A main.go from the project root directory launches both the app and the agent. you can run it with

```bash
go run main.go
```

Alternatively, you can run the two programs separately (not recommended, as the helper script initializes some shared systems):

```bash
# Terminal 1
go run cmd/app/main.go

# Terminal 2
go run cmd/agent/
```

## Testing

### Tests

Unit and integration tests can be run for the entire project by running.

WARNING: Running tests will delete the database

```bash
go test ./...
```

### Testing the API

You can test the API with a program like CURL. Commands are in the [API Documentation](#api-endpoints)

## Documentation

### Application (app)

Purpose:

* Hosts frontend
* Hosts REST API
* Processes expressions into an AST for later calculation by the agent
* Manages and serves tasks
* Saves to and loads from a sqlite database

Features:

* Logging
* In-depth error handling
* Unit and integration tests

#### API Endpoints

##### /api/v1/register
