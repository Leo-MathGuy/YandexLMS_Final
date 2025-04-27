# Yandex LMS Final Project

## Testing from console

### Register

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"login":"bob","password":"123"}' \
  localhost:8080/api/v1/register
```

### Login

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"login":"bob","password":"123"}' \
  localhost:8080/api/v1/login
```
