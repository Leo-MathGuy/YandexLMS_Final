# Yandex Go Course Final Project

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Leo-MathGuy/YandexLMS_Final/go.yml?label=tests)](https://github.com/Leo-MathGuy/YandexLMS_Final/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/Leo-MathGuy/YandexLMS_Final/badge.svg?branch=main)](https://coveralls.io/github/Leo-MathGuy/YandexLMS_Final?branch=main)

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

Unit, integration, race condition and coverage tests can be run for the entire project by running the following command:

WARNING: Running tests will delete the database

```bash
go test -cover ./... -race
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
* Serves a gRPC server for the agent

Features:

* Logging
* In-depth error handling
* Unit and integration tests

#### API Endpoints

##### /api/v1/register
