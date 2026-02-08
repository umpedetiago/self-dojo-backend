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

## Endpoints

## Swagger (OpenAPI)

- Swagger UI: `http://localhost:8080/swagger`
- Spec: `http://localhost:8080/openapi.yaml`

### POST `/auth/register`

Body:

```json
{ "email": "user@example.com", "password": "minimo-8-caracteres" }
```

### POST `/auth/login`

Body:

```json
{ "email": "user@example.com", "password": "minimo-8-caracteres" }
```

### GET `/me`

Header:

- `Authorization: Bearer <token>`

