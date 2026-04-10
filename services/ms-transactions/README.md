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
| `JWT_SECRET` | Sim | Deve coincidir com o **ms-users** para aceitar o mesmo `access_token` (equivale ao `ILIACHALLENGE` do enunciado). |
| `JWT_INTERNAL_SECRET` | Sim | Segredo **distinto** de `JWT_SECRET`, partilhado com o ms-users para JWT só nas rotas internas (`ILIACHALLENGE_INTERNAL` no enunciado). |
| `USERS_SERVICE_BASE_URL` | Sim | URL base do ms-users (ex.: `http://localhost:3002` no host; em Compose ver [README na raiz](../../README.md#autenticação-dupla-e-comunicação-interna)). |
| `USERS_SERVICE_TIMEOUT` | Não | Timeout do cliente HTTP para a verificação de utilizador ativo (omissão: `5s`). |
| `OPENAPI_SPEC` | Não | Caminho do OpenAPI; vazio desativa Swagger UI. |

### Comunicação interna (ms-transactions)

- **`GET /internal/wallet/{userId}/balance`** — Protegida por Bearer com JWT assinado com `JWT_INTERNAL_SECRET` e audience esperada. Resposta **200** com JSON `{"balance": <int64>}` (unidades mínimas, mesma lógica que `GET /balance` público). Destina-se ao ms-users antes do soft delete (fluxo B no [README da raiz](../../README.md#autenticação-dupla-e-comunicação-interna)).

Antes de persistir uma nova linha em **`POST /transactions`**, o serviço chama o ms-users em **`GET /internal/users/{userId}`**. Utilizador inexistente ou apagado (soft delete) → **403 Forbidden**; falha de rede ou 5xx no ms-users → **503 Service Unavailable**. O corpo continua a exigir `user_id` igual ao `sub` do JWT do utilizador.

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

## Idempotência (`POST /transactions`)

O header opcional **`Idempotency-Key`** evita criar transações duplicadas quando o cliente repete o mesmo `POST` (timeout, retry de rede, etc.).

- **Escopo:** par (`usuário do JWT`, valor do header). Outro usuário pode usar a mesma string de chave sem colidir.
- **Comportamento:** mesma chave **e** mesmo corpo efetivo (`user_id`, `type`, `amount`) → resposta **201** de novo com o **mesmo** `id` e JSON (não há segunda linha na carteira). Mesma chave com corpo **diferente** → **409 Conflict**.
- **Opcional:** clientes antigos ou fluxos sem retry podem omitir o header; cada chamada continua criando uma transação nova. Isso segue o padrão de APIs de pagamento públicas que **recomendam** chave para retry mas não obrigam.
- **Recomendado:** gerar um UUID (ou outro id estável) **por ação do usuário** e reutilizá-lo em todos os retries daquela ação.
- **Limite:** no máximo **255** code points Unicode; acima disso a API responde **400** (`idempotency key exceeds maximum length`).

Exemplo com chave:

```bash
curl -sS -X POST http://localhost:3001/transactions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"user_id":"<mesmo UUID do sub>","type":"CREDIT","amount":100}'
```

Detalhes e textos de erro: [`ms-transactions.yaml`](ms-transactions.yaml) (parâmetro `Idempotency-Key` em `POST /transactions`).

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
