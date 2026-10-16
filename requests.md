# Тестовые запросы к API (ПЗ №3)

Сервер: `go run ./cmd/server` (или `make run`), адрес `http://localhost:8080`.
Для Windows PowerShell — вариант с `Invoke-RestMethod` под каждым запросом.

## 1. Проверка состояния — `GET /health` → 200

```bash
curl -i http://localhost:8080/health
```
```powershell
Invoke-RestMethod http://localhost:8080/health
```
Ответ: `{"status":"ok"}`

## 2. Создание задачи — `POST /tasks` → 201

```bash
curl -i -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Купить молоко"}'
```
```powershell
Invoke-RestMethod -Method POST http://localhost:8080/tasks `
  -ContentType "application/json; charset=utf-8" `
  -Body '{"title":"Купить молоко"}'
```
Ответ: `{"id":1,"title":"Купить молоко","done":false}`, заголовок `Location: /tasks/1`

## 3. Список задач — `GET /tasks` → 200

```bash
curl -i http://localhost:8080/tasks
```
```powershell
Invoke-RestMethod http://localhost:8080/tasks
```

## 4. Фильтр — `GET /tasks?q=молоко` → 200

```bash
curl -i -G http://localhost:8080/tasks --data-urlencode "q=молоко"
```
```powershell
Invoke-RestMethod "http://localhost:8080/tasks?q=молоко"
```
Поиск без учёта регистра по подстроке в `title`.

## 5. Одна задача — `GET /tasks/1` → 200

```bash
curl -i http://localhost:8080/tasks/1
```
```powershell
Invoke-RestMethod http://localhost:8080/tasks/1
```

## 6. Ошибки валидации

```bash
# нет title → 400 Bad Request
curl -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{}'

# битый JSON → 400 Bad Request
curl -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":'

# title короче 3 символов → 422 Unprocessable Entity
curl -i -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"ab"}'

# неверный Content-Type → 400 Bad Request
curl -i -X POST http://localhost:8080/tasks -H "Content-Type: text/plain" -d 'hello'
```

## 7. Ошибки в пути

```bash
# id не число → 400 Bad Request
curl -i http://localhost:8080/tasks/abc

# задачи нет → 404 Not Found
curl -i http://localhost:8080/tasks/9999
```

## 8. Изменение флага done — `PATCH /tasks/2` → 200

```bash
curl -i -X PATCH http://localhost:8080/tasks/2 \
  -H "Content-Type: application/json" \
  -d '{"done":true}'
```
```powershell
Invoke-RestMethod -Method PATCH http://localhost:8080/tasks/2 `
  -ContentType "application/json" -Body '{"done":true}'
```

## 9. Удаление — `DELETE /tasks/1` → 204

```bash
curl -i -X DELETE http://localhost:8080/tasks/1
```
```powershell
Invoke-RestMethod -Method DELETE http://localhost:8080/tasks/1
```

## 10. CORS preflight — `OPTIONS /tasks` → 204

```bash
curl -i -X OPTIONS http://localhost:8080/tasks -H "Origin: http://localhost:3000"
```
Ответ содержит `Access-Control-Allow-Origin: *` и список разрешённых методов.
