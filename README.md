# ília — desafio (Go)

Repositório com a solução em **Go** para o desafio da ília Digital. O enunciado original está em [`challenge.md`](challenge.md) (referência Node.js; aqui os microsserviços são implementados em Go quando aplicável).

## Estrutura do repositório

| Caminho | Descrição |
|---------|-----------|
| [`go.mod`](go.mod) | Módulo Go único na raiz (`ilia-golang-challenge`); **sem** `go.work`. |
| [`services/ms-users/`](services/ms-users/) | Microsserviço de usuários (HTTP, Postgres, migrações embutidas). |
| [`services/ms-users/README.md`](services/ms-users/README.md) | Setup e uso do **ms-users**. |
| [`services/ms-transactions/`](services/ms-transactions/) | Microsserviço de carteira / transações (HTTP, Postgres dedicado, JWT). |
| [`services/ms-transactions/README.md`](services/ms-transactions/README.md) | Setup e uso do **ms-transactions**. |
| [`services/ms-transactions/ms-transactions.yaml`](services/ms-transactions/ms-transactions.yaml) | Especificação OpenAPI da carteira. |

## Pré-requisitos gerais

- **Go** compatível com a versão indicada em [`go.mod`](go.mod) (toolchain do módulo na raiz).
- **Docker** e **Docker Compose** (opcional, para subir Postgres + API em cada serviço).
- **Make** (opcional; `Makefile` em `services/ms-users` e `services/ms-transactions`).

## Início rápido

1. Microsserviço de usuários: [`services/ms-users/README.md`](services/ms-users/README.md).
2. Microsserviço de carteira: [`services/ms-transactions/README.md`](services/ms-transactions/README.md).
3. Escopo do desafio: [`challenge.md`](challenge.md).

## Go (a partir da raiz do repositório)

```bash
go test ./...
go build -o bin/ms-users-api ./services/ms-users/cmd/api
go build -o bin/ms-transactions-api ./services/ms-transactions/cmd/api
```

Imagens Docker (contexto = raiz do repo):

```bash
docker build -f services/ms-users/Dockerfile .
docker build -f services/ms-transactions/Dockerfile .
```
