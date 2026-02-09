# self-dojo-backend

## Rodar o banco (Postgres)

```bash
docker compose up -d
```

Se o container do Postgres falhar com erro de "data format" ou "pg_ctlcluster", apague a pasta `data` (dados antigos), depois suba de novo: `docker compose up -d`.  
Se a tabela `users` já existir com `id` inteiro (versão antiga), apague-a para usar UUIDv7: `DROP TABLE IF EXISTS users;` no Postgres e reinicie a API.

Banco padrão (do `docker-compose.yml`):

- user: `postgres`
- password: `postgres`
- db: `go_db`
- port: `5433` (host; dentro do container continua 5432)

## Rodar a API

Defina variáveis de ambiente (veja `.env.example`):

- `DATABASE_URL` (ex.: `postgres://postgres:postgres@localhost:5433/go_db?sslmode=disable`)
- `JWT_SECRET`
- `PORT` (default `8080`)

Rodar:

```bash
go run ./cmd
```

## Endpoints (URLs exatas – 404 = path ou método errado)

- `GET  http://localhost:8080/ping`
- `GET  http://localhost:8080/swagger`
- `GET  http://localhost:8080/openapi.yaml`
- `POST http://localhost:8080/auth/register`
- `POST http://localhost:8080/auth/login`
- `POST http://localhost:8080/auth/logout` (requer `Authorization: Bearer <token>`)
- `POST http://localhost:8080/auth/forgot-password`
- `POST http://localhost:8080/auth/reset-password`
- `GET  http://localhost:8080/me` (requer header `Authorization: Bearer <token>`)

Não existe prefixo `/api` – use `/auth/register`, não `/api/auth/register`.

## Swagger (OpenAPI)

- Swagger UI: `http://localhost:8080/swagger`
- Spec: `http://localhost:8080/openapi.yaml`

### POST `/auth/register`

Body:

```json
{ "username": "meu_usuario", "email": "user@example.com", "password": "minimo-8-caracteres" }
```

### POST `/auth/login`

Body:

```json
{ "email": "user@example.com", "password": "minimo-8-caracteres" }
```

### POST `/auth/forgot-password`

Solicita reset de senha. Em desenvolvimento, retorna o token na resposta (em produção, enviaria por email).

Body:

```json
{ "email": "user@example.com" }
```

Resposta (dev):

```json
{
  "message": "if the email exists, a reset link will be sent",
  "token": "abc123...",
  "expires_at": "2026-02-08T16:00:00Z"
}
```

### POST `/auth/reset-password`

Resetar senha usando o token recebido.

Body:

```json
{ "token": "abc123...", "password": "nova-senha-minimo-8" }
```

### GET `/me`

Retorna o perfil do usuário autenticado: `id`, `username`, `email`.

Header:

- `Authorization: Bearer <token>`

