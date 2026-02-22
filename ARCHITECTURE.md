# TinyStock - Production Architecture

## Overview

TinyStock is a full-stack stock tracking application with a layered backend (Go/Gin) and Streamlit frontend (Python). The system supports multi-user authentication, per-user watchlists and portfolios, and production-grade middleware.

## Architecture Diagram

```mermaid
flowchart TB
    subgraph Client [Client Layer]
        Streamlit[Streamlit Frontend]
    end

    subgraph Backend [Go Backend]
        subgraph Routes [routes/]
            AuthRoutes[Auth Routes]
            StockRoutes[Stock Routes]
            WatchlistRoutes[Watchlist Routes]
            PortfolioRoutes[Portfolio Routes]
        end

        subgraph Middleware [middleware/]
            Logger[Request Logger]
            Recovery[Error Recovery]
            CORS[CORS]
            RateLimit[Rate Limiter]
            Auth[JWT Auth]
        end

        subgraph Handlers [handlers/]
            AuthHandler[Auth Handler]
            StockHandler[Stock Handler]
            WatchlistHandler[Watchlist Handler]
            PortfolioHandler[Portfolio Handler]
        end

        subgraph Services [services/]
            AuthService[Auth Service]
            StockService[Stock Service + Cache]
            WatchlistService[Watchlist Service]
            PortfolioService[Portfolio Service]
        end

        subgraph Repository [repository/]
            UserRepo[User Repository]
            WatchlistRepo[Watchlist Repository]
            PortfolioRepo[Portfolio Repository]
        end
    end

    subgraph Data [Data Layer]
        DB[(PostgreSQL / SQLite)]
    end

    subgraph External [External]
        YahooAPI[Yahoo Finance API]
    end

    Streamlit -->|HTTP + JWT| Routes
    Routes --> Middleware
    Middleware --> Handlers
    Handlers --> Services
    Services --> Repository
    Services --> YahooAPI
    Repository --> DB
```

## Folder Structure

```
TinyStock/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point
│   ├── config/
│   │   └── config.go                # Environment config
│   ├── models/
│   │   ├── user.go
│   │   ├── watchlist.go
│   │   ├── portfolio.go
│   │   └── stock.go
│   ├── repository/
│   │   ├── repository.go            # DB interface
│   │   ├── user_repository.go
│   │   ├── watchlist_repository.go
│   │   ├── portfolio_repository.go
│   │   └── postgres/
│   │       └── postgres.go          # Postgres implementation
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── stock_service.go         # Yahoo proxy + cache
│   │   ├── watchlist_service.go
│   │   └── portfolio_service.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── stock_handler.go
│   │   ├── watchlist_handler.go
│   │   └── portfolio_handler.go
│   ├── routes/
│   │   └── routes.go                # Route registration
│   ├── middleware/
│   │   ├── logger.go
│   │   ├── recovery.go
│   │   ├── cors.go
│   │   ├── ratelimit.go
│   │   └── auth.go
│   ├── internal/
│   │   └── response/                 # Structured responses
│   │       └── response.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── frontend/
│   ├── app.py
│   ├── api_client.py
│   ├── views/
│   │   ├── dashboard.py
│   │   ├── quote.py
│   │   ├── watchlist.py
│   │   └── portfolio.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── docker-compose.yml
├── .env.example
└── README.md
```

## Layer Responsibilities

| Layer | Responsibility |
|-------|----------------|
| **routes/** | HTTP route registration, groups public vs protected |
| **handlers/** | Parse request, validate input, call service, return response |
| **services/** | Business logic, orchestration, external API calls |
| **repository/** | Database CRUD, abstraction over Postgres/SQLite |
| **middleware/** | Cross-cutting: logging, auth, rate limit, CORS |
| **config/** | Load env vars, validate config at startup |

## Data Flow

1. **Request** -> Middleware (logger, CORS, rate limit, auth for protected routes)
2. **Handler** -> Extract params, validate, call Service
3. **Service** -> Business logic, call Repository and/or Stock API
4. **Repository** -> Execute SQL, return domain models
5. **Response** -> Structured JSON with status, data, error

## Database Schema

```sql
-- users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- watchlist (per-user)
CREATE TABLE watchlist (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    UNIQUE(user_id, symbol)
);

-- holdings (per-user)
CREATE TABLE holdings (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(18,6) NOT NULL,
    buy_price DECIMAL(18,6) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/auth/register | No | Register user |
| POST | /api/auth/login | No | Login, returns JWT |
| GET | /api/quote/:symbol | No | Stock quote (cached) |
| GET | /api/history/:symbol | No | 30-day history |
| GET | /api/search | No | Symbol search |
| GET | /api/watchlist | Yes | User watchlist |
| POST | /api/watchlist | Yes | Add to watchlist |
| DELETE | /api/watchlist/:symbol | Yes | Remove from watchlist |
| GET | /api/portfolio | Yes | User portfolio with P&L |
| POST | /api/portfolio | Yes | Add holding |
| DELETE | /api/portfolio/:id | Yes | Remove holding |

## Environment Variables

```
# Backend
PORT=8080
DB_DRIVER=postgres|sqlite
DATABASE_URL=postgres://user:pass@host:5432/tinystock
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h
RATE_LIMIT=100
CORS_ORIGINS=http://localhost:8501

# Frontend
TINYSTOCK_API_URL=http://localhost:8080
```
