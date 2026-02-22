package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"tinystock/backend/models"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// DB implements repository.DB for SQLite
type DB struct {
	conn *sql.DB
}

// New creates a new SQLite database connection
func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
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
	if err := d.seedIfEmpty(); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}
	return d, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS watchlist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(user_id, symbol)
		)`,
		`CREATE TABLE IF NOT EXISTS holdings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			quantity REAL NOT NULL,
			buy_price REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range queries {
		if _, err := d.conn.Exec(q); err != nil {
			return err
		}
	}
	if err := d.migrateLegacyUserScopedTables(); err != nil {
		return err
	}
	return nil
}

// migrateLegacyUserScopedTables upgrades older single-user schemas that did not include user_id.
func (d *DB) migrateLegacyUserScopedTables() error {
	watchlistHasUserID, err := d.tableHasColumn("watchlist", "user_id")
	if err != nil {
		return err
	}
	holdingsHasUserID, err := d.tableHasColumn("holdings", "user_id")
	if err != nil {
		return err
	}
	if watchlistHasUserID && holdingsHasUserID {
		return nil
	}

	defaultUserID, err := d.ensureDefaultUserID()
	if err != nil {
		return err
	}

	if !watchlistHasUserID {
		if err := d.migrateLegacyWatchlist(defaultUserID); err != nil {
			return err
		}
	}
	if !holdingsHasUserID {
		if err := d.migrateLegacyHoldings(defaultUserID); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) tableHasColumn(table, column string) (bool, error) {
	rows, err := d.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid      int
			name     string
			colType  string
			notNull  int
			defaultV sql.NullString
			primaryK int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &primaryK); err != nil {
			return false, err
		}
		if strings.EqualFold(name, column) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (d *DB) ensureDefaultUserID() (string, error) {
	var userID string
	err := d.conn.QueryRow("SELECT id FROM users ORDER BY created_at ASC LIMIT 1").Scan(&userID)
	if err == nil {
		return userID, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// Legacy DB might have watchlist/holdings data but no users. Create a demo owner to preserve data.
	userID = uuid.New().String()
	hash, err := hashPassword("demo123")
	if err != nil {
		return "", err
	}
	if _, err := d.conn.Exec(
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID,
		"demo@tinystock.app",
		hash,
	); err != nil {
		return "", err
	}
	return userID, nil
}

func (d *DB) migrateLegacyWatchlist(defaultUserID string) error {
	tx, err := d.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DROP TABLE IF EXISTS watchlist_legacy"); err != nil {
		return err
	}
	if _, err := tx.Exec("ALTER TABLE watchlist RENAME TO watchlist_legacy"); err != nil {
		return err
	}
	if _, err := tx.Exec(`CREATE TABLE watchlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		UNIQUE(user_id, symbol)
	)`); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`INSERT OR IGNORE INTO watchlist (user_id, symbol)
		 SELECT ?, UPPER(TRIM(symbol))
		 FROM watchlist_legacy
		 WHERE symbol IS NOT NULL AND TRIM(symbol) <> ''`,
		defaultUserID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE watchlist_legacy"); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) migrateLegacyHoldings(defaultUserID string) error {
	tx, err := d.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DROP TABLE IF EXISTS holdings_legacy"); err != nil {
		return err
	}
	if _, err := tx.Exec("ALTER TABLE holdings RENAME TO holdings_legacy"); err != nil {
		return err
	}
	if _, err := tx.Exec(`CREATE TABLE holdings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL,
		buy_price REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`INSERT INTO holdings (user_id, symbol, quantity, buy_price)
		 SELECT ?, UPPER(TRIM(symbol)), quantity, buy_price
		 FROM holdings_legacy
		 WHERE symbol IS NOT NULL AND TRIM(symbol) <> ''`,
		defaultUserID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec("DROP TABLE holdings_legacy"); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) seedIfEmpty() error {
	var count int
	if err := d.conn.QueryRow("SELECT COUNT(*) FROM watchlist").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// Create demo user for development
	demoID := uuid.New().String()
	hash, _ := hashPassword("demo123")
	_, _ = d.conn.Exec("INSERT OR IGNORE INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		demoID, "demo@tinystock.app", hash)
	// Seed watchlist for demo user
	for _, s := range []string{"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA"} {
		_, _ = d.conn.Exec("INSERT OR IGNORE INTO watchlist (user_id, symbol) VALUES (?, ?)", demoID, s)
	}
	for _, h := range []struct {
		sym        string
		qty, price float64
	}{
		{"AAPL", 10, 175.50}, {"MSFT", 5, 380.00}, {"GOOGL", 3, 140.00},
	} {
		_, _ = d.conn.Exec("INSERT INTO holdings (user_id, symbol, quantity, buy_price) VALUES (?, ?, ?, ?)",
			demoID, h.sym, h.qty, h.price)
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
	_, err = d.conn.ExecContext(ctx, "INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		user.ID, user.Email, hash)
	return err
}

// GetByEmail implements UserRepository
func (d *DB) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := d.conn.QueryRowContext(ctx, "SELECT id, email, password_hash FROM users WHERE email = ?", email).
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
	err := d.conn.QueryRowContext(ctx, "SELECT id, email, password_hash FROM users WHERE id = ?", id).
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
	_, err := d.conn.ExecContext(ctx, "INSERT INTO watchlist (user_id, symbol) VALUES (?, ?)", userID, symbol)
	return err
}

// Remove implements WatchlistRepository
func (d *DB) Remove(ctx context.Context, userID, symbol string) error {
	_, err := d.conn.ExecContext(ctx, "DELETE FROM watchlist WHERE user_id = ? AND symbol = ?", userID, symbol)
	return err
}

// List implements WatchlistRepository
func (d *DB) List(ctx context.Context, userID string) ([]models.WatchlistItem, error) {
	rows, err := d.conn.QueryContext(ctx, "SELECT id, symbol FROM watchlist WHERE user_id = ? ORDER BY symbol", userID)
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

// Add implements PortfolioRepository
func (d *DB) AddHolding(ctx context.Context, userID string, symbol string, quantity, buyPrice float64) error {
	_, err := d.conn.ExecContext(ctx, "INSERT INTO holdings (user_id, symbol, quantity, buy_price) VALUES (?, ?, ?, ?)",
		userID, symbol, quantity, buyPrice)
	return err
}

// Remove implements PortfolioRepository
func (d *DB) RemoveHolding(ctx context.Context, userID string, id int64) error {
	_, err := d.conn.ExecContext(ctx, "DELETE FROM holdings WHERE user_id = ? AND id = ?", userID, id)
	return err
}

// List implements PortfolioRepository
func (d *DB) ListHoldings(ctx context.Context, userID string) ([]models.Holding, error) {
	rows, err := d.conn.QueryContext(ctx, "SELECT id, symbol, quantity, buy_price FROM holdings WHERE user_id = ? ORDER BY symbol", userID)
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
