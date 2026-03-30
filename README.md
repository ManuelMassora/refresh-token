# Refresh Token API

A REST API built in **Go** that demonstrates the implementation of **access token** and **refresh token** authentication flow using JWT.

## About

This project is a practical study on stateless authentication using the dual-token strategy:

- **Access Token** — short-lived token used to authenticate requests.
- **Refresh Token** — long-lived token used to issue new access tokens without requiring the user to log in again.

## Tech Stack

| Technology | Purpose |
|---|---|
| [Go](https://golang.org/) | Main language |
| [Chi](https://github.com/go-chi/chi) | HTTP router |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | JWT generation and validation |
| [GORM](https://gorm.io/) | ORM for database access |
| [PostgreSQL](https://www.postgresql.org/) | Relational database |
| [godotenv](https://github.com/joho/godotenv) | Environment variable loading |
| [google/uuid](https://github.com/google/uuid) | UUID generation |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Password hashing |

## Project Structure

```
refresh-token/
├── internal/
│   ├── config/     # Environment configuration loader
│   ├── db/         # Database initialization
│   ├── di/         # Dependency injection container
│   └── route/      # Route registration
├── main.go         # Application entry point
├── go.mod
├── go.sum
└── .env            # Environment variables (not committed)
```

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [PostgreSQL](https://www.postgresql.org/)

## Getting Started

**1. Clone the repository**

```bash
git clone https://github.com/ManuelMassora/refresh-token.git
cd refresh-token
```

**2. Set up environment variables**

Create a `.env` file at the project root:

```env
DSN=host=localhost user=postgres password=yourpassword dbname=refresh_token_db port=5432 sslmode=disable
SERVER_PORT=8080
JWT_SECRET=your_jwt_secret_key
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
```

**3. Install dependencies**

```bash
go mod tidy
```

**4. Run the application**

```bash
go run main.go
```

The server will start on the port defined in `SERVER_PORT`.

## How It Works

```
Client                    Server
  |                          |
  |-- POST /login ---------->|  (credentials)
  |<-- access_token + -------|
  |    refresh_token         |
  |                          |
  |-- GET /protected ------->|  (Authorization: Bearer <access_token>)
  |<-- 200 OK ---------------|
  |                          |
  |  [access_token expires]  |
  |                          |
  |-- POST /refresh -------->|  (refresh_token)
  |<-- new access_token -----|
```

## Dependencies

```
github.com/go-chi/chi/v5
github.com/golang-jwt/jwt/v5
github.com/google/uuid
github.com/joho/godotenv
golang.org/x/crypto
gorm.io/driver/postgres
gorm.io/gorm
```

## License

This project is intended for educational and study purposes.

---

Made by [Manuel Massora](https://github.com/ManuelMassora)

---

# API de Refresh Token

Uma API REST construída em **Go** que demonstra a implementação do fluxo de autenticação com **access token** e **refresh token** usando JWT.

## Sobre

Este projeto é um estudo prático sobre autenticação stateless utilizando a estratégia de duplo token:

- **Access Token** — token de curta duração utilizado para autenticar requisições.
- **Refresh Token** — token de longa duração utilizado para emitir novos access tokens sem que o utilizador precise fazer login novamente.

## Tecnologias

| Tecnologia | Finalidade |
|---|---|
| [Go](https://golang.org/) | Linguagem principal |
| [Chi](https://github.com/go-chi/chi) | Roteador HTTP |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | Geração e validação de JWT |
| [GORM](https://gorm.io/) | ORM para acesso ao banco de dados |
| [PostgreSQL](https://www.postgresql.org/) | Banco de dados relacional |
| [godotenv](https://github.com/joho/godotenv) | Carregamento de variáveis de ambiente |
| [google/uuid](https://github.com/google/uuid) | Geração de UUIDs |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Hash de senhas |

## Estrutura do Projeto

```
refresh-token/
├── internal/
│   ├── config/     # Carregamento de configurações do ambiente
│   ├── db/         # Inicialização do banco de dados
│   ├── di/         # Container de injeção de dependências
│   └── route/      # Registro de rotas
├── main.go         # Ponto de entrada da aplicação
├── go.mod
├── go.sum
└── .env            # Variáveis de ambiente (não versionado)
```

## Pré-requisitos

- [Go 1.21+](https://golang.org/dl/)
- [PostgreSQL](https://www.postgresql.org/)

## Como Executar

**1. Clone o repositório**

```bash
git clone https://github.com/ManuelMassora/refresh-token.git
cd refresh-token
```

**2. Configure as variáveis de ambiente**

Crie um arquivo `.env` na raiz do projeto:

```env
DSN=host=localhost user=postgres password=suasenha dbname=refresh_token_db port=5432 sslmode=disable
SERVER_PORT=8080
JWT_SECRET=sua_chave_secreta_jwt
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
```

**3. Instale as dependências**

```bash
go mod tidy
```

**4. Execute a aplicação**

```bash
go run main.go
```

O servidor irá iniciar na porta definida em `SERVER_PORT`.

## Como Funciona

```
Cliente                   Servidor
  |                          |
  |-- POST /login ---------->|  (credenciais)
  |<-- access_token + -------|
  |    refresh_token         |
  |                          |
  |-- GET /protegido ------->|  (Authorization: Bearer <access_token>)
  |<-- 200 OK ---------------|
  |                          |
  |  [access_token expira]   |
  |                          |
  |-- POST /refresh -------->|  (refresh_token)
  |<-- novo access_token ----|
```

## Dependências

```
github.com/go-chi/chi/v5
github.com/golang-jwt/jwt/v5
github.com/google/uuid
github.com/joho/godotenv
golang.org/x/crypto
gorm.io/driver/postgres
gorm.io/gorm
```

## Licença

Este projeto é destinado a fins educacionais e de estudo.

---

Feito por [Manuel Massora](https://github.com/ManuelMassora)
