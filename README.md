# TinyStock

A production-ready full-stack stock tracking application with **Python (Streamlit)** frontend and **Go (Gin)** backend. Built as a resume project for stock/finance industry roles.

## Features

- **JWT Authentication** - User registration and login
- **Per-User Data** - Watchlist and portfolio scoped to each user
- **Stock Quote Lookup** - Search symbols, live prices, 30-day charts
- **Watchlist** - Add/remove stocks, view at a glance
- **Portfolio Tracking** - Real-time P&L, total value, return percentage
- **Production Stack** - PostgreSQL, rate limiting, CORS, structured errors

## Tech Stack

| Layer | Technology |
|-------|-------------|
| Frontend | Python 3.11+, Streamlit, requests |
| Backend | Go 1.21+, Gin, layered architecture |
| Database | PostgreSQL (prod), SQLite (dev) |
| Stock Data | Yahoo Finance (proxied through backend only) |

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for full design.

```
backend/
├── config/         # Environment config
├── models/         # Domain models
├── repository/     # DB layer (SQLite + Postgres)
├── services/       # Business logic, stock API, auth
├── handlers/       # HTTP handlers
├── routes/         # Route registration
├── middleware/     # Logger, CORS, rate limit, auth
└── internal/       # Shared utilities
```

## Quick Start

### Prerequisites

- Go 1.21+
- Python 3.11+
- (Optional) Docker & Docker Compose

### Run Locally (SQLite)

**1. Backend**

```bash
cd backend
go mod tidy
go run .
```

**2. Frontend**

```bash
cd frontend
pip install -r requirements.txt
streamlit run app.py
```

**3. Demo Login**

- Email: `demo@tinystock.app`
- Password: `demo123`

### Run with Docker (PostgreSQL)

```bash
docker-compose up --build
```

- Backend: http://localhost:8080
- Frontend: http://localhost:8501
- Postgres: localhost:5432

### Development (SQLite, no Postgres)

```bash
docker-compose -f docker-compose.dev.yml up --build
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/auth/register` | No | Register user |
| POST | `/api/auth/login` | No | Login, returns JWT |
| GET | `/api/quote/:symbol` | No | Stock quote |
| GET | `/api/history/:symbol` | No | 30-day history |
| GET | `/api/search?q=` | No | Symbol search |
| GET | `/api/watchlist` | Yes | User watchlist |
| POST | `/api/watchlist` | Yes | Add to watchlist |
| DELETE | `/api/watchlist/:symbol` | Yes | Remove |
| GET | `/api/portfolio` | Yes | Portfolio with P&L |
| POST | `/api/portfolio` | Yes | Add holding |
| DELETE | `/api/portfolio/:id` | Yes | Remove holding |

Protected endpoints require `Authorization: Bearer <token>` header.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | API port |
| DB_DRIVER | sqlite | sqlite or postgres |
| DATABASE_URL | ./tinystock.db | DB connection string |
| JWT_SECRET | (dev) | JWT signing key |
| JWT_EXPIRY | 24h | Token expiry |
| RATE_LIMIT | 100 | Requests per minute |
| CORS_ORIGINS | * | Allowed origins |
| TINYSTOCK_API_URL | http://localhost:8080 | Backend URL (frontend) |

## License

MIT
