FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Migrations are embedded via //go:embed in internal/infrastructure/db.
RUN CGO_ENABLED=0 go build -o /out/api ./services/ms-users/cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api .
COPY services/ms-users/ms-users.yaml /app/ms-users.yaml
ENV OPENAPI_SPEC=/app/ms-users.yaml
# PORT comes from the runtime environment (docker-compose, `docker run -e`, etc.).
# Default 3002 is applied in Go when PORT is unset.
EXPOSE 3002
CMD ["./api"]
