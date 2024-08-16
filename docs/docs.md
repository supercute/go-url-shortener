
# Документация (generated from chatgpt)

## API

---

**Создание сокращенной ссылки**

**Запрос:**

```
POST /create
Content-Type: application/json

{
  "link": "http://example.com",
  "name": "shortlink"
}
```

**Ответ:**

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "short_link": "shortlink"
}
```

**Ошибки:**

- `400 Bad Request`: Неверный формат запроса или отсутствует обязательное поле.
- `409 Conflict`: Имя ссылки уже существует.

---

**Редирект по сокращенной ссылке**

**Запрос:**

```
GET /{shortName}
```

**Ответ:**

```
HTTP/1.1 301 Moved Permanently
Location: http://example.com
```

**Ошибки:**

- `404 Not Found`: Ссылка не найдена.

---

**Создание токена**

**Запрос:**

```
POST /register
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Ответ:**

```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "token": "your_jwt_token"
}
```

**Ошибки:**

- `400 Bad Request`: Неверный формат запроса или отсутствует обязательное поле.
- `409 Conflict`: Email уже используется.

---

**Удаление пользователя**

**Запрос:**

```
POST /admin/delete
Content-Type: application/json
Authorization: Bearer admin_token

{
  "email": "user@example.com"
}
```

**Ответ:**

```
HTTP/1.1 204 No Content
```

**Ошибки:**

- `403 Forbidden`: Неавторизованный доступ.
- `400 Bad Request`: Неверный формат запроса или отсутствует обязательное поле.
- `404 Not Found`: Пользователь не найден.

---

**Получение списка пользователей**

**Запрос:**

```
GET /admin/users
Authorization: Bearer admin_token
```

**Ответ:**

```
HTTP/1.1 200 OK
Content-Type: application/json

{
    "user@example.com": 5,
    "user2@example.com": 10
}
```

**Ошибки:**

- `403 Forbidden`: Неавторизованный доступ.
