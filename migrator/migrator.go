package migrator

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/pressly/goose/v3"
)

const (
	DialectPostgres = "postgres"
	DialectSQLite3  = "sqlite3"
)

type Option func(*Migrator)

func WithLogger(logger *slog.Logger) Option {
	return func(m *Migrator) {
		if logger != nil {
			m.logger = logger
		}
	}
}

func WithDialect(dialect string) Option {
	return func(m *Migrator) {
		m.dialect = dialect
	}
}

func WithPath(path string) Option {
	return func(m *Migrator) {
		m.path = path
	}
}

type Migrator struct {
	logger  *slog.Logger
	dialect string
	path    string
}

func NewMigrator(opts ...Option) *Migrator {
	m := &Migrator{
		logger:  slog.New(slog.DiscardHandler),
		dialect: DialectPostgres,
		path:    ".",
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *Migrator) UpContext(ctx context.Context, db *sql.DB) error {
	goose.SetLogger(NewGooseLogger(m.logger))

	if err := goose.SetDialect(m.dialect); err != nil {
		return err
	}

	if err := goose.UpContext(ctx, db, m.path); err != nil {
		if !errors.Is(err, goose.ErrNoMigrations) && !errors.Is(err, goose.ErrNoMigrationFiles) {
			return err
		}
	}

	return nil
}

func (m *Migrator) ResetContext(ctx context.Context, db *sql.DB) error {
	goose.SetLogger(NewGooseLogger(m.logger))

	if err := goose.SetDialect(m.dialect); err != nil {
		return err
	}

	if err := goose.ResetContext(ctx, db, m.path); err != nil {
		if !errors.Is(err, goose.ErrNoMigrations) && !errors.Is(err, goose.ErrNoMigrationFiles) {
			return err
		}
	}

	return nil
}
