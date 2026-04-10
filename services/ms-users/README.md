# ms-users

Microsserviço HTTP de **cadastro de usuários**: validação de domínio, hash de senha com **bcrypt**, persistência em **PostgreSQL** e migrações aplicadas na subida da aplicação.

## Pré-requisitos

- Go (versão conforme o `go.mod` / `toolchain` na **raiz do repositório**; não há `go.mod` neste diretório).
- PostgreSQL acessível pela `DATABASE_URL`, **ou** Docker Compose (este diretório sobe Postgres + API).

## Configuração

1. Copie o exemplo de ambiente (recomendado para desenvolvimento local):

   ```bash
   cp .env.example .env
   ```

2. Ajuste `DATABASE_URL` se o Postgres não for o padrão do exemplo.

Variáveis principais (detalhes e opcionais em [`.env.example`](.env.example)):

| Variável | Obrigatória | Descrição |
|----------|-------------|-----------|
| `PORT` | Sim | Porta HTTP (ex.: `3002`). |
| `DATABASE_URL` | Sim | URL Postgres (driver `pgx` via `database/sql`). |
| `JWT_SECRET` | Sim | Segredo HS256 para tokens de **utilizador** (login); deve coincidir com o ms-transactions para o mesmo `access_token`. Equivale ao `ILIACHALLENGE` do enunciado. |
| `JWT_INTERNAL_SECRET` | Sim | Segredo **distinto** de `JWT_SECRET`, partilhado com o ms-transactions só para JWT entre serviços (`ILIACHALLENGE_INTERNAL` no enunciado). |
| `TRANSACTIONS_SERVICE_BASE_URL` | Sim | URL base do ms-transactions (ex.: `http://localhost:3001` no host; em Compose ver [README na raiz](../../README.md#autenticação-dupla-e-comunicação-interna)). |
| `TRANSACTIONS_SERVICE_TIMEOUT` | Não | Timeout do cliente HTTP para a verificação de saldo antes do delete (omissão: `5s`). |
| `OPENAPI_SPEC` | Não | Caminho do arquivo OpenAPI; vazio desativa Swagger UI. |
| `DOTENV_FILE` | Não | Arquivo carregado pelo `godotenv` (o `make run` define; ver abaixo). |

### Comunicação interna (ms-users)

- **`GET /internal/users/{id}`** — Protegida por Bearer com JWT assinado com `JWT_INTERNAL_SECRET` e audience esperada pelo serviço. Resposta **200** com JSON mínimo (`id`, `active: true`) se o utilizador existe e não está soft-deleted; **404** caso contrário. Não expõe email nem password. Destina-se ao ms-transactions (fluxo A no [README da raiz](../../README.md#autenticação-dupla-e-comunicação-interna)).

Antes de **`DELETE /users/{id}`** (soft delete), o serviço consulta o ms-transactions em **`GET /internal/wallet/{userId}/balance`**. Se o saldo consolidado (unidades mínimas) for **zero**, o delete prossegue (**204**). Saldo não zero → **409 Conflict**; indisponibilidade do ms-transactions → **503**.

## Executar localmente (sem Docker)

Com Postgres já rodando e `.env` criado:

```bash
cd services/ms-users
make run
```

Outros perfis de arquivo:

```bash
make run-example   # DOTENV_FILE=.env.example
make run-dev       # DOTENV_FILE=.env.develop
make run-prod      # DOTENV_FILE=.env.prod
```

Equivalente manual:

```bash
# A partir da raiz do repositório (usa o `go.mod` da raiz):
export DOTENV_FILE="$(pwd)/services/ms-users/.env"
go run ./services/ms-users/cmd/api
```

## Docker Compose (Postgres + API)

Neste diretório, o Compose usa contexto de build na **raiz do repo** (`context: ../..`, `dockerfile: services/ms-users/Dockerfile`) e `--env-file` explícito:

| Comando | Arquivo de variáveis (substituição no `docker-compose.yml`) |
|---------|--------------------------------------------------------------|
| `make up` ou `make up-dev` | `.env` (o `setup` cria a partir de `.env.example` se faltar) |
| `make up-example` | `.env.example` |
| `make up-prod` | `.env.prod` (`setup-prod` se necessário) |

```bash
make up-dev      # ou: make up (alias)
make up-example
make up-prod
```

- API: porta **3002** por padrão (ou `PORT` no arquivo usado pelo `up-*`).
- Postgres: usuário/senha/db `users`, porta publicada **5432**.

Encerrar:

```bash
make down
```

Logs e status:

```bash
make logs
make ps
```

## API e documentação

- **Base local:** `http://localhost:3002` (ajuste se mudar `PORT`).
- **Criar usuário:** `POST /users` — corpo JSON com `first_name`, `last_name`, `email`, `password` (ver contrato em [`ms-users.yaml`](ms-users.yaml)).
- **Listar / obter / atualizar:** `GET /users`, `GET /users/{id}`, `PATCH /users/{id}` — apenas usuários **ativos** (`deleted_at` nulo).
- **Remover:** `DELETE /users/{id}` — **exclusão lógica** (`deleted_at` preenchido); o e-mail pode ser reutilizado em novo cadastro.
- **OpenAPI em YAML:** [`ms-users.yaml`](ms-users.yaml).
- **Swagger UI** (se `OPENAPI_SPEC` apontar para o arquivo): `GET /swagger/` e `GET /openapi.yaml`.

## Testes e qualidade

```bash
make test   # go test ./... -count=1
make fmt    # go fmt ./...
make vet    # go vet ./...
```

Liste todos os alvos do `Makefile`:

```bash
make help
```

## Layout do código (resumo)

- `cmd/api` — ponto de entrada, carrega config, DB, migrações e HTTP.
- `internal/domain` — entidades e regras de validação.
- `internal/application/.../usecase` — casos de uso.
- `internal/infrastructure/httpapi` — handlers e roteamento.
- `internal/infrastructure/persistence` — repositório Postgres.
- `internal/infrastructure/db` — migrações SQL embutidas.

## Imagem Docker

O `Dockerfile` neste diretório espera build com contexto na **raiz do repositório** (por exemplo `docker build -f services/ms-users/Dockerfile .` ou `make up-dev`). Ele compila o binário estático e copia o `ms-users.yaml` para uso com `OPENAPI_SPEC` no container. O `docker-compose.yml` define as variáveis necessárias para subir o serviço junto ao Postgres.
