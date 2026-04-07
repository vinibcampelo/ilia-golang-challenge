# ília — desafio (Go)

Repositório com a solução em **Go** para o desafio da ília Digital. O enunciado original está em [`challenge.md`](challenge.md) (referência Node.js; aqui os microsserviços são implementados em Go quando aplicável).

## Estrutura do repositório

| Caminho | Descrição |
|---------|-----------|
| [`go.mod`](go.mod) | Módulo Go único na raiz (`ilia-golang-challenge`); **sem** `go.work`. |
| [`services/ms-users/`](services/ms-users/) | Microsserviço de usuários (HTTP, Postgres, migrações embutidas). |
| [`services/ms-users/README.md`](services/ms-users/README.md) | Setup e uso do **ms-users**. |
| [`ms-transactions/ms-transactions.yaml`](ms-transactions/ms-transactions.yaml) | Especificação OpenAPI do serviço de transações (carteira); implementação do serviço pode ser adicionada nesta pasta. |

## Pré-requisitos gerais

- **Go** compatível com a versão indicada em [`go.mod`](go.mod) (toolchain do módulo na raiz).
- **Docker** e **Docker Compose** (opcional, para subir Postgres + API via compose em `services/ms-users`).
- **Make** (opcional; o `Makefile` de `services/ms-users` automatiza tarefas comuns).

## Início rápido

1. Microsserviço de usuários: siga [`services/ms-users/README.md`](services/ms-users/README.md).
2. Leia o escopo completo em [`challenge.md`](challenge.md).

## Go (a partir da raiz do repositório)

```bash
go test ./...
go build -o bin/ms-users-api ./services/ms-users/cmd/api
```

Imagem Docker do ms-users (contexto = raiz do repo):

```bash
docker build -f services/ms-users/Dockerfile .
```
