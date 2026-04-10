# ília — desafio (Go)

Repositório com a solução em **Go** para o desafio da ília Digital: dois microsserviços HTTP com PostgreSQL (**usuários** e **carteira / transações**). O enunciado original está em [`challenge.md`](challenge.md) (referência em Node.js; aqui a implementação é em Go).

A estrutura deste documento segue uma ordem habitual em READMEs de software: **o que precisa na máquina** → **como subir o ambiente** → **configuração e ferramentas** → **estrutura do repositório** e **ligações** para documentação detalhada.

## Índice

- [Pré-requisitos](#pré-requisitos)
- [Início rápido (Docker)](#início-rápido-docker)
- [Portas e stacks](#portas-e-stacks-valores-por-omissão)
- [Configuração e JWT](#configuração-e-jwt)
- [Autenticação dupla e comunicação interna](#autenticação-dupla-e-comunicação-interna)
- [Makefile na raiz](#makefile-na-raiz)
- [Docker e Make por serviço](#docker-e-make-por-serviço)
- [Desenvolvimento sem Docker (API no host)](#desenvolvimento-sem-docker-api-no-host)
- [Go (módulo na raiz)](#go-módulo-na-raiz)
- [Imagens Docker (build manual)](#imagens-docker-build-manual)
- [Estrutura do repositório](#estrutura-do-repositório)
- [Documentação e contratos](#documentação-e-contratos)

---

## Pré-requisitos

### Inicializar o projeto em Docker (fluxo principal)

Para **subir Postgres e as duas APIs em contentores** (`make setup && make up` ou o equivalente com `docker compose` em cada serviço), **não precisa de Go instalado na máquina**: a compilação acontece **dentro da imagem** no build do Docker.

O que precisa na máquina é:

| Ferramenta | Para quê |
|------------|----------|
| **Docker** | Motor de contentores (Docker Engine ou Docker Desktop, conforme o seu SO) |
| **Docker Compose v2** | Comando `docker compose` usado pelos `Makefile` (no Docker Desktop costuma vir incluído) |
| **GNU Make** | Para os comandos `make` deste README na raiz e em `services/*` (muitas vezes já instalado em Linux/macOS; em Windows, use WSL ou instale Make) |
| **Bash** | Os `Makefile` usam `SHELL := bash` (hábitual em Linux, macOS e WSL) |

Em resumo: o **requisito mínimo para subir o stack** é **Docker com Compose**; **Make** (e Bash) entram porque a documentação usa `make` como atalho. Se preferir só Docker, pode copiar os `.env` manualmente e **executar** `docker compose` em `services/ms-users` e `services/ms-transactions` como nos READMEs de cada serviço.

### Versão de Go (compilar ou testar no host)

A versão usada pelo código está em [`go.mod`](go.mod):

| Declaração | Valor |
|------------|--------|
| Diretiva `go` | **1.26** |
| `toolchain` | **go1.26.1** |

Isto só é necessário se for **compilar ou testar fora do Docker** (`go test`, `go build`, `make build-api`, `make run-users`, etc.). Com Go 1.21+, o comando `go` pode descarregar automaticamente a toolchain do `go.mod`.

### Só API no host (sem Docker)

Precisa de **Go** (conforme tabela acima), **Make**, **Bash** e **PostgreSQL** acessível pelos `.env`; fluxo completo em [Desenvolvimento sem Docker (API no host)](#desenvolvimento-sem-docker-api-no-host).

---

## Início rápido (Docker)

Na **raiz** do repositório:

```bash
make setup
make up
```

O `setup` cria `services/ms-users/.env` e `services/ms-transactions/.env` a partir dos respetivos `.env.example` se ainda não existirem. O `up` sobe **primeiro** o stack do **ms-users**, depois o do **ms-transactions** (cada um com o seu Postgres).

**Confirmar que o ambiente está ativo:** estado dos contentores com `make ps`. Por omissão as APIs escutam nas portas da tabela abaixo; se `OPENAPI_SPEC` estiver definido em cada `.env`, o Swagger fica em `/swagger/` em cada serviço.

**Fluxo típico para testar as duas APIs**

1. **ms-users** — criar utilizador (`POST /users`) e obter token (`POST /auth`). Detalhes em [`services/ms-users/README.md`](services/ms-users/README.md).
2. **ms-transactions** — chamar rotas autenticadas com `Authorization: Bearer <access_token>`. Detalhes em [`services/ms-transactions/README.md`](services/ms-transactions/README.md).

O token é emitido pelo ms-users; o ms-transactions valida **JWT HS256** com o mesmo segredo (ver secção seguinte).

**Parar os stacks (mantém volumes / dados do Postgres):**

```bash
make down
```

**Apagar também os volumes (reinicia bases de dados):**

```bash
make down-v
```

---

## Portas e stacks (valores por omissão)

Cada serviço tem o seu `docker-compose.yml` sob `services/<nome>/`. O nome do projeto Compose corresponde à pasta (`ms-users`, `ms-transactions`).

| Serviço | API HTTP (host) | PostgreSQL (host) | Base de dados (omissão) |
|---------|-----------------|-------------------|-------------------------|
| **ms-users** | [http://localhost:3002](http://localhost:3002) | `localhost:5432` | `users` |
| **ms-transactions** | [http://localhost:3001](http://localhost:3001) | `localhost:5433` | `transactions` |

A porta **5433** no host para transações evita colisão com o Postgres do ms-users na **5432**.

---

## Configuração e JWT

Cada API usa o seu ficheiro **`.env`** dentro da pasta do serviço (`services/ms-users/.env`, `services/ms-transactions/.env`). Variáveis e exemplos completos: [`services/ms-users/.env.example`](services/ms-users/.env.example) e [`services/ms-transactions/.env.example`](services/ms-transactions/.env.example).

**Obrigatório para o fluxo users → carteira:** o valor de **`JWT_SECRET`** tem de ser **idêntico** nos dois `.env`. Caso contrário, o `access_token` devolvido por `POST /auth` no ms-users não será aceite no ms-transactions. O valor por omissão nos exemplos é o mesmo (`change-me-in-production`); altere em **ambos** os ficheiros se mudar num deles.

**Obrigatório para chamadas internas entre APIs:** **`JWT_INTERNAL_SECRET`** (mesmo valor nos dois serviços, **distinto** de `JWT_SECRET`), mais **`TRANSACTIONS_SERVICE_BASE_URL`** no ms-users e **`USERS_SERVICE_BASE_URL`** no ms-transactions. Detalhes em [Autenticação dupla e comunicação interna](#autenticação-dupla-e-comunicação-interna).

---

## Autenticação dupla e comunicação interna

O [`challenge.md`](challenge.md) distingue a segurança **exposta a clientes** da segurança **entre microsserviços**. Neste repositório a correspondência é:

| Enunciado (challenge) | Variável aqui | Uso |
|------------------------|---------------|-----|
| `ILIACHALLENGE` | `JWT_SECRET` | JWT do **utilizador** (login no ms-users; `Authorization: Bearer` nas rotas públicas). |
| `ILIACHALLENGE_INTERNAL` | `JWT_INTERNAL_SECRET` | JWT **curto**, emitido por cada serviço ao chamar o outro; **só** nas rotas `GET /internal/...`. Não reutilizar o access token do utilizador. |

### Regras de negócio (REST entre serviços)

| Fluxo | Quem chama quem | Rota interna | Efeito na API pública |
|-------|-----------------|--------------|------------------------|
| **A** — criar transação | ms-transactions → ms-users | `GET /internal/users/{id}` | Utilizador inexistente ou soft-deleted → **403** no `POST /transactions`. Falha de rede / 5xx no ms-users → **503**. |
| **B** — apagar conta | ms-users → ms-transactions | `GET /internal/wallet/{userId}/balance` | Resposta JSON com campo `balance` (inteiro, unidades mínimas). Saldo **0** → `DELETE /users/{id}` pode concluir (**204**). Saldo ≠ 0 → **409**. Falha no ms-transactions → **503**. |

A definição de “saldo zero” é a **mesma** que a do saldo público (soma das transações em unidades mínimas).

As rotas `/internal/...` **não** constam do OpenAPI público; destinam-se a tráfego entre serviços na mesma rede de confiança (por exemplo Docker ou `localhost` em desenvolvimento).

### Docker: duas stacks Compose

Os `docker-compose.yml` ligam as APIs à rede externa partilhada **`ilia-challenge-internal`** (criada automaticamente em `make up` / `make -C services/… up`). Os contentores têm nomes estáveis **`ilia-ms-users`** e **`ilia-ms-transactions`**. As URLs HTTP entre APIs no Compose estão **fixas no YAML** (`http://ilia-ms-transactions:3001` e `http://ilia-ms-users:3002`). O **`PORT` dentro do contentor** da API também é fixo (**3001** / **3002**) para coincidir com essas URLs; a porta publicada no host usa **`MS_TRANSACTIONS_HOST_PORT`** e **`MS_USERS_HOST_PORT`** (omissão 3001 / 3002), para não haver dessincronia quando o `.env` trazia outro `PORT`. Com **`make run`** no host, o `.env` com `PORT` e `localhost` para o outro serviço continua a aplicar-se só ao processo no host.

---

## Makefile na raiz

Na raiz existe um [`Makefile`](Makefile) que **orquestra** os dois serviços (chama `make -C services/ms-users …` e `make -C services/ms-transactions …`). Não há `docker-compose.yml` na raiz.

Liste todos os alvos e variáveis úteis:

```bash
make help
```

| Alvo | Descrição |
|------|-----------|
| `setup` | `setup` em ambos os serviços (`.env` a partir de `.env.example` se faltar) |
| `setup-dev` / `setup-prod` | Equivalente em ambos os serviços |
| `up` / `up-dev` | Sobe os dois stacks (variáveis Compose a partir do `.env` de cada serviço) |
| `up-example` | Sobe os dois stacks só com `.env.example` de cada serviço |
| `up-prod` | Sobe os dois stacks com `.env.prod` de cada serviço |
| `down` / `down-v` | Para os stacks (com ou sem remoção de volumes) |
| `build` / `restart` / `ps` | Imagens, reinício e estado dos contentores (ambos os projetos) |
| `logs-users` / `logs-transactions` | Logs em seguimento de um stack de cada vez |
| `run-users` / `run-transactions` | Inicia a API no **host** (Go), não no Docker, com o `.env` desse serviço |
| `test` | `go test ./... -count=1` a partir da raiz |
| `fmt` / `vet` | Formatação e `go vet` em todo o módulo |
| `build-api` | Compila `bin/ms-users-api` e `bin/ms-transactions-api` |

**Apenas um serviço:** use o Makefile dentro da pasta, por exemplo:

```bash
make -C services/ms-users help
make -C services/ms-users up
```

O mesmo padrão aplica-se a `services/ms-transactions`.

---

## Docker e Make por serviço

O ficheiro **`docker-compose.yml`** e o **`Makefile`** de cada API estão em:

- [`services/ms-users/`](services/ms-users/)
- [`services/ms-transactions/`](services/ms-transactions/)

O build das imagens usa **contexto na raiz do repositório** (`context: ../..` nos compose, ver Dockerfiles em cada serviço). O Makefile da raiz apenas delega; a configuração detalhada de Compose e de ambiente continua documentada em cada README de serviço.

---

## Desenvolvimento sem Docker (API no host)

É preciso PostgreSQL acessível conforme `DATABASE_URL` em cada `.env` (no host, as portas por omissão são **5432** para users e **5433** para transactions).

A partir da **raiz**:

```bash
make run-users
make run-transactions
```

Em janelas separadas, ou siga os equivalentes com `DOTENV_FILE` e `go run` descritos em [`services/ms-users/README.md`](services/ms-users/README.md) e [`services/ms-transactions/README.md`](services/ms-transactions/README.md).

---

## Go (módulo na raiz)

O módulo único está em [`go.mod`](go.mod) (**Go 1.26**, toolchain **go1.26.1**; **sem** `go.work`). Detalhes e requisitos de instalação: [Pré-requisitos](#pré-requisitos).

```bash
go test ./...
go fmt ./...
go vet ./...
```

Ou na raiz:

```bash
make test
make fmt
make vet
make build-api
```

Binários gerados por `make build-api`:

- `bin/ms-users-api`
- `bin/ms-transactions-api`

---

## Imagens Docker (build manual)

O contexto de build é sempre a **raiz** do repositório:

```bash
docker build -f services/ms-users/Dockerfile .
docker build -f services/ms-transactions/Dockerfile .
```

Isto alinha-se com o que o `docker compose` de cada serviço utiliza.

---

## Estrutura do repositório

| Caminho | Descrição |
|---------|-----------|
| [`go.mod`](go.mod) | Módulo Go `ilia-golang-challenge` na raiz |
| [`Makefile`](Makefile) | Orquestração: ambos os serviços + alvos Go do módulo completo |
| [`services/ms-users/`](services/ms-users/) | Microsserviço de utilizadores (HTTP, Postgres, migrações) |
| [`services/ms-users/Makefile`](services/ms-users/Makefile) | Alvos locais (Compose, `run`, testes só deste serviço) |
| [`services/ms-users/docker-compose.yml`](services/ms-users/docker-compose.yml) | Postgres + API ms-users |
| [`services/ms-users/ms-users.yaml`](services/ms-users/ms-users.yaml) | OpenAPI (Swagger) do ms-users |
| [`services/ms-users/README.md`](services/ms-users/README.md) | Setup, rotas e detalhes do ms-users |
| [`services/ms-transactions/`](services/ms-transactions/) | Microsserviço de carteira (HTTP, Postgres, JWT) |
| [`services/ms-transactions/Makefile`](services/ms-transactions/Makefile) | Alvos locais deste serviço |
| [`services/ms-transactions/docker-compose.yml`](services/ms-transactions/docker-compose.yml) | Postgres + API ms-transactions |
| [`services/ms-transactions/ms-transactions.yaml`](services/ms-transactions/ms-transactions.yaml) | OpenAPI da carteira |
| [`services/ms-transactions/README.md`](services/ms-transactions/README.md) | Setup, JWT e exemplos `curl` |
| [`challenge.md`](challenge.md) | Enunciado do desafio |

---

## Documentação e contratos

- Enunciado: [`challenge.md`](challenge.md)
- MS Users (API, Swagger, testes): [`services/ms-users/README.md`](services/ms-users/README.md) · OpenAPI [`services/ms-users/ms-users.yaml`](services/ms-users/ms-users.yaml)
- MS Transactions (API, JWT, Swagger): [`services/ms-transactions/README.md`](services/ms-transactions/README.md) · OpenAPI [`services/ms-transactions/ms-transactions.yaml`](services/ms-transactions/ms-transactions.yaml)
