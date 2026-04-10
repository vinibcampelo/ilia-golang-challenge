# ília — desafio (Go)

Solução em **Go** ao desafio da ília Digital ([`challenge.md`](challenge.md) — referência em Node.js). São **dois microsserviços HTTP**, cada um com **PostgreSQL** próprio:

| Serviço | Porta omissão | O que faz |
|---------|---------------|-----------|
| **ms-users** | 3002 | Utilizadores: registo, `POST /auth` (JWT), CRUD; soft delete. |
| **ms-transactions** | 3001 | Carteira: transações (crédito/débito), saldo; aceita o **mesmo** `JWT_SECRET` que o users para validar o token do cliente. |

**Integração entre serviços (Parte 2):** comunicação **REST** com segurança **distinta** da API pública — JWT interno de curta duração (`JWT_INTERNAL_SECRET`, alinhado a `ILIACHALLENGE_INTERNAL` no enunciado). O **transactions** chama o **users** antes de criar transação; o **users** chama o **transactions** antes de apagar conta (saldo zero). Rotas: `GET /internal/users/{id}` e `GET /internal/wallet/{userId}/balance`. Variáveis, erros HTTP e exemplos: [`services/ms-users/README.md`](services/ms-users/README.md) e [`services/ms-transactions/README.md`](services/ms-transactions/README.md).

---

## Índice

- [Início rápido (Docker)](#início-rápido-docker)
- [Pré-requisitos](#pré-requisitos)
- [Portas](#portas)
- [Segredos JWT e rede interna](#segredos-jwt-e-rede-interna)
- [Makefile na raiz](#makefile-na-raiz)
- [Sem Docker · Go · build de imagens](#sem-docker--go--build-de-imagens)
- [Estrutura e documentação](#estrutura-e-documentação)

---

## Início rápido (Docker)

Na **raiz**:

```bash
make setup   # cria .env a partir de .env.example em cada serviço, se faltar
make up      # sobe ms-users, depois ms-transactions (cada um com Postgres)
```

- Estado dos contentores: `make ps` · parar: `make down` · apagar dados: `make down-v`.
- Testar: criar utilizador e login no **users**; usar o token no **transactions** (passos em cada README). Swagger em `/swagger/` se `OPENAPI_SPEC` estiver definido no `.env`.

**Requisito mínimo:** Docker com **Compose v2**; **Make** e Bash são atalhos (pode usar só `docker compose` nas pastas `services/*`). **Go não é obrigatório** para subir o stack — a build corre na imagem.

---

## Pré-requisitos

| Cenário | Precisa de |
|---------|------------|
| Só Docker | Docker, Compose v2, Make (recomendado), Bash |
| Testar/compilar no host | Go conforme [`go.mod`](go.mod) (1.26 / toolchain 1.26.1), Postgres acessível pelos `.env` |

---

## Portas

| Serviço | API (host) | Postgres (host) |
|---------|------------|-----------------|
| ms-users | [localhost:3002](http://localhost:3002) | 5432 |
| ms-transactions | [localhost:3001](http://localhost:3001) | 5433 |

A porta **5433** evita colisão com o Postgres do users na 5432.

---

## Segredos JWT e rede interna

- **`JWT_SECRET`** — igual nos dois `.env`; token do **utilizador** (equiv. `ILIACHALLENGE`). Sem isto o login do users não vale no transactions.
- **`JWT_INTERNAL_SECRET`** — igual nos dois, **diferente** de `JWT_SECRET`; só para chamadas **serviço a serviço** nas rotas `/internal/...`.

Em **Docker**, os dois `docker-compose` ligam-se à rede **`ilia-challenge-internal`** (criada pelo `make up`), com contentores `ilia-ms-users` e `ilia-ms-transactions`; URLs e portas **dentro** do mesh estão fixas no YAML para não confundir com `localhost` do `.env`. No **host** (`make run`), usa-se `localhost` e `PORT` como nos `.env.example`. Mais pormenor: READMEs dos serviços e [`services/ms-users/.env.example`](services/ms-users/.env.example) / [`services/ms-transactions/.env.example`](services/ms-transactions/.env.example).

---

## Makefile na raiz

Orquestra os dois serviços (`make -C services/ms-users …` e idem para transactions). Lista completa:

```bash
make help
```

Alvos frequentes: `setup`, `up`, `down`, `down-v`, `ps`, `test`, `run-users`, `run-transactions`, `build-api`. Um serviço só: `make -C services/ms-users help` (ou `ms-transactions`).

---

## Sem Docker · Go · build de imagens

- **API no host:** Postgres acessível; `make run-users` e `make run-transactions` na raiz (ou `go run` com `DOTENV_FILE` — ver READMEs dos serviços).
- **Testes / formato:** `make test`, `make fmt`, `make vet` ou `go test ./...` na raiz.
- **Binários:** `make build-api` → `bin/ms-users-api`, `bin/ms-transactions-api`.
- **Imagens (contexto = raiz):** `docker build -f services/ms-users/Dockerfile .` e o equivalente para `ms-transactions`.

---

## Estrutura e documentação

| Caminho | Conteúdo |
|---------|-----------|
| [`go.mod`](go.mod) | Módulo único na raiz |
| [`Makefile`](Makefile) | Orquestração + alvos Go |
| [`services/ms-users/`](services/ms-users/) | Código, [`docker-compose.yml`](services/ms-users/docker-compose.yml), [`README.md`](services/ms-users/README.md), [`ms-users.yaml`](services/ms-users/ms-users.yaml) |
| [`services/ms-transactions/`](services/ms-transactions/) | Idem para a carteira |

- Enunciado: [`challenge.md`](challenge.md)
