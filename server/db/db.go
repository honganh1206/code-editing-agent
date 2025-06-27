package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	Dsn          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func OpenDB(cfg Config, schema string) (*sql.DB, error) {
	dbDir := filepath.Dir(cfg.Dsn)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		err = os.MkdirAll(dbDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", cfg.Dsn)
	if err != nil {
		return nil, err
	}

	if err = initializeSchema(db, schema); err != nil {
		db.Close()
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	duration, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify connection to db is still alive
	err = db.PingContext(ctx)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func initializeSchema(db *sql.DB, schema string) error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema initialization SQL: %w", err)
	}
	return nil
}
