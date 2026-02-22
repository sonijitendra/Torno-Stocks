package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"tinystock/backend/models"
)

// DB implements repository.DB for PostgreSQL
type DB struct {
	conn *sql.DB
}

// New creates a new PostgreSQL database connection
func New(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (d *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS watchlist (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			symbol VARCHAR(20) NOT NULL,
			UNIQUE(user_id, symbol)
		)`,
		`CREATE TABLE IF NOT EXISTS holdings (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			symbol VARCHAR(20) NOT NULL,
			quantity DECIMAL(18,6) NOT NULL,
			buy_price DECIMAL(18,6) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
	}
	for _, q := range queries {
		if _, err := d.conn.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// Create implements UserRepository
func (d *DB) Create(ctx context.Context, user *models.User) error {
	user.ID = uuid.New().String()
	hash, err := hashPassword(user.PasswordHash)
	if err != nil {
		return err
	}
	_, err = d.conn.ExecContext(ctx, "INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
		user.ID, user.Email, hash)
	return err
}

// GetByEmail implements UserRepository
func (d *DB) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := d.conn.QueryRowContext(ctx, "SELECT id::text, email, password_hash FROM users WHERE email = $1", email).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID implements UserRepository
func (d *DB) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := d.conn.QueryRowContext(ctx, "SELECT id::text, email, password_hash FROM users WHERE id = $1", id).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Add implements WatchlistRepository
func (d *DB) Add(ctx context.Context, userID, symbol string) error {
	_, err := d.conn.ExecContext(ctx, "INSERT INTO watchlist (user_id, symbol) VALUES ($1, $2)", userID, symbol)
	return err
}

// Remove implements WatchlistRepository
func (d *DB) Remove(ctx context.Context, userID, symbol string) error {
	_, err := d.conn.ExecContext(ctx, "DELETE FROM watchlist WHERE user_id = $1 AND symbol = $2", userID, symbol)
	return err
}

// List implements WatchlistRepository
func (d *DB) List(ctx context.Context, userID string) ([]models.WatchlistItem, error) {
	rows, err := d.conn.QueryContext(ctx, "SELECT id, symbol FROM watchlist WHERE user_id = $1 ORDER BY symbol", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.WatchlistItem
	for rows.Next() {
		var w models.WatchlistItem
		w.UserID = userID
		if err := rows.Scan(&w.ID, &w.Symbol); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

// AddHolding implements PortfolioRepository
func (d *DB) AddHolding(ctx context.Context, userID string, symbol string, quantity, buyPrice float64) error {
	_, err := d.conn.ExecContext(ctx, "INSERT INTO holdings (user_id, symbol, quantity, buy_price) VALUES ($1, $2, $3, $4)",
		userID, symbol, quantity, buyPrice)
	return err
}

// RemoveHolding implements PortfolioRepository
func (d *DB) RemoveHolding(ctx context.Context, userID string, id int64) error {
	_, err := d.conn.ExecContext(ctx, "DELETE FROM holdings WHERE user_id = $1 AND id = $2", userID, id)
	return err
}

// ListHoldings implements PortfolioRepository
func (d *DB) ListHoldings(ctx context.Context, userID string) ([]models.Holding, error) {
	rows, err := d.conn.QueryContext(ctx, "SELECT id, symbol, quantity, buy_price FROM holdings WHERE user_id = $1 ORDER BY symbol", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var holdings []models.Holding
	for rows.Next() {
		var h models.Holding
		h.UserID = userID
		if err := rows.Scan(&h.ID, &h.Symbol, &h.Quantity, &h.BuyPrice); err != nil {
			return nil, err
		}
		holdings = append(holdings, h)
	}
	return holdings, rows.Err()
}
