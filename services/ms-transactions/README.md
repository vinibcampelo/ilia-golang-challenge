# ms-transactions

Microsserviço HTTP de **carteira (wallet)**: créditos e débitos por usuário, saldo consolidado via SQL (`SUM` + `CASE`), PostgreSQL dedicado e **JWT HS256** nas rotas. O token é emitido pelo **ms-users** (`POST /auth`); o **`JWT_SECRET` deve ser o mesmo** nos dois serviços para validar o `sub` (UUID do usuário).

Débitos concorrentes para o mesmo usuário usam **`pg_advisory_xact_lock`** + transação para não permitir saldo negativo.

Montantes na API e no banco são **inteiros em unidades mínimas da moeda** (ex.: centavos para BRL), sem ponto flutuante.

## Pré-requisitos

- Go (versão conforme o `go.mod` na **raiz do repositório**).
- PostgreSQL acessível pela `DATABASE_URL`, **ou** Docker Compose neste diretório (Postgres + API).

## Configuração

1. `cp .env.example .env`
2. Ajuste `DATABASE_URL` se necessário. O exemplo usa **`localhost:5433`** no host para não colidir com o Postgres padrão do **ms-users** (`5432`).

| Variável | Obrigatória | Descrição |
|----------|-------------|-----------|
| `PORT` | Sim | Porta HTTP (ex.: **3001**). |
| `DATABASE_URL` | Sim | URL Postgres (`pgx` / `database/sql`). |
| `JWT_SECRET` | Sim | Deve coincidir com o **ms-users** para aceitar o mesmo `access_token`. |
| `OPENAPI_SPEC` | Não | Caminho do OpenAPI; vazio desativa Swagger UI. |

## Executar localmente (sem Docker)

Com Postgres da carteira acessível e `.env` criado:

```bash
cd services/ms-transactions
make run
```

Equivalente a partir da raiz:

```bash
export DOTENV_FILE="$(pwd)/services/ms-transactions/.env"
go run ./services/ms-transactions/cmd/api
```

## Docker Compose

Build com contexto na **raiz do repositório** (`dockerfile: services/ms-transactions/Dockerfile`).

```bash
cd services/ms-transactions
make up-dev
```

- API: **3001** por padrão.
- Postgres: DB `transactions`, porta publicada **5433** no host (configurável via `POSTGRES_PORT`).

## Obter JWT e chamar a API

1. Suba o **ms-users**, crie usuário e faça login (ver [README do ms-users](../ms-users/README.md)).
2. Use o `access_token` retornado:

```bash
TOKEN="<access_token>"

curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:3001/balance

curl -sS -X POST http://localhost:3001/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<mesmo UUID do sub>","type":"CREDIT","amount":100}'
```

O corpo de `POST /transactions` exige `user_id` **igual** ao `sub` do JWT; caso contrário a API responde **403**.

## OpenAPI

Especificação: [`ms-transactions.yaml`](ms-transactions.yaml). Com `OPENAPI_SPEC` definido, a UI fica em `/swagger/`.

## Testes

```bash
# A partir da raiz
go test ./services/ms-transactions/... -count=1
```

Teste de **concorrência** com Postgres real (opcional):

```bash
export MS_TRANSACTIONS_INTEGRATION_DATABASE_URL='postgres://transactions:transactions@localhost:5433/transactions?sslmode=disable'
go test ./services/ms-transactions/internal/infrastructure/persistence/... -count=1 -run Integration
```

## Imagem Docker (raiz do repo)

```bash
docker build -f services/ms-transactions/Dockerfile .
```
