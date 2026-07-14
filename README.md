# Soccer Manager API

A RESTful API for managing fantasy football teams. Users register, receive a squad of 20 players, and can buy or sell players on a transfer market.

## Database Schema

![Database Schema](docs/schema.png)

---

## Tech Stack

| Layer            | Tool                         |
| ---------------- | ---------------------------- |
| Language         | Go 1.26.1                    |
| HTTP framework   | Gin                          |
| Database         | PostgreSQL 18 (via Docker)   |
| Driver / pool    | pgx/v5 + pgxpool             |
| Query generation | sqlc                         |
| Migrations       | golang-migrate               |
| Auth             | JWT (golang-jwt/v5) + bcrypt |
| Localisation     | go-i18n (English + Georgian) |
| Rate limiting    | golang.org/x/time/rate       |
| API docs         | swaggo/swag (OpenAPI 3)      |

---

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) and Docker Compose

---

## Setup

```bash
git clone https://github.com/faikbairamov/soccer-manager.git
cd soccer-manager

cp .env.example .env        # review and adjust if needed

make docker-up              # start PostgreSQL in Docker
make migrate-up             # apply all schema migrations
make run                    # start the API on :8080
```

The server is ready when you see `"server starting"` in the logs.

---

## Architecture

The project follows a layered architecture so business rules stay independent from transport and persistence details.

```text
Client (Postman / Swagger / Frontend)
		  |
		  v
Gin Router + Middleware
(request id, i18n, rate limit, auth)
		  |
		  v
Handlers
(HTTP parsing, validation, response mapping)
		  |
		  v
Services
(business rules, transfer logic, ownership checks)
		  |
		  v
Repository (sqlc store + transactions)
		  |
		  v
PostgreSQL
```

### Layer Responsibilities

- Router and middleware: cross-cutting concerns such as authentication, rate limiting, and localisation.
- Handlers: translate HTTP requests into service calls and map domain errors to API responses.
- Services: enforce domain rules (single team per user, transfer constraints, budget updates, value bump logic).
- Repository: execute SQL through sqlc-generated methods and wrap multi-step writes in database transactions.

### Why This Helps

- Testability: business logic can be tested without HTTP or database wiring.
- Maintainability: each layer has one job, making changes safer and easier to review.
- Data consistency: transfer operations are atomic, so partial updates cannot be committed.

---

## Endpoints

### Public

| Method | Path                  | Description                                |
| ------ | --------------------- | ------------------------------------------ |
| `GET`  | `/health`             | Health check — 200 if DB is up, 503 if not |
| `GET`  | `/swagger/index.html` | Swagger UI                                 |

### Auth (rate-limited: 5 req/s, burst 10)

| Method | Path                    | Description                                    |
| ------ | ----------------------- | ---------------------------------------------- |
| `POST` | `/api/v1/auth/register` | Create account + auto-generate 20-player squad |
| `POST` | `/api/v1/auth/login`    | Authenticate and receive a JWT                 |

### Teams (JWT required)

| Method  | Path               | Description                         |
| ------- | ------------------ | ----------------------------------- |
| `GET`   | `/api/v1/teams/me` | Get your team and total squad value |
| `PATCH` | `/api/v1/teams/me` | Update team name and/or country     |

### Players (JWT required)

| Method  | Path                  | Description                                                      |
| ------- | --------------------- | ---------------------------------------------------------------- |
| `GET`   | `/api/v1/players/:id` | Get a player's details                                           |
| `PATCH` | `/api/v1/players/:id` | Update a player's first name, last name, or country (owner only) |

### Transfer Market (JWT required)

| Method   | Path                        | Description                                                |
| -------- | --------------------------- | ---------------------------------------------------------- |
| `GET`    | `/api/v1/transfers`         | List all players on the market (`?page=1&limit=20`)        |
| `POST`   | `/api/v1/transfers`         | Put one of your players on the market with an asking price |
| `DELETE` | `/api/v1/transfers/:id`     | Remove your player from the market                         |
| `POST`   | `/api/v1/transfers/:id/buy` | Purchase a listed player                                   |

---

## Swagger UI

With the server running, open:

```
http://localhost:8080/swagger/index.html
```

You can authorise with a JWT token and try every endpoint directly from the browser. The Postman collection is available at `docs/soccer-manager.postman_collection.json`.

---

## Localisation

Send `Accept-Language: ka` to receive error messages in Georgian. Omit the header or send `Accept-Language: en` for English.

---
