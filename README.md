# ília — desafio (Go)

Repositório com a solução em **Go** para o desafio da ília Digital. O enunciado original está em [`challenge.md`](challenge.md) (referência Node.js; aqui os microsserviços são implementados em Go quando aplicável).

## Estrutura do repositório

| Caminho | Descrição |
|---------|-----------|
| [`ms-users/`](ms-users/) | Microsserviço de usuários (HTTP, Postgres, migrações embutidas). |
| [`ms-users/README.md`](ms-users/README.md) | Setup e uso do **ms-users**. |
| [`ms-transactions/ms-transactions.yaml`](ms-transactions/ms-transactions.yaml) | Especificação OpenAPI do serviço de transações (carteira); implementação do serviço pode ser adicionada nesta pasta. |

## Pré-requisitos gerais

- **Go** compatível com a versão indicada em `ms-users/go.mod` (toolchain do módulo).
- **Docker** e **Docker Compose** (opcional, para subir Postgres + API via compose no `ms-users`).
- **Make** (opcional; o `Makefile` do `ms-users` automatiza tarefas comuns).

## Início rápido

1. Microsserviço de usuários: siga [`ms-users/README.md`](ms-users/README.md).
2. Leia o escopo completo em [`challenge.md`](challenge.md).