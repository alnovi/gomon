package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type key string

const txKey key = "tx"

type Option func(c *Client) error

func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) error {
		if logger != nil {
			c.logger = logger
		}
		return nil
	}
}

type Client struct {
	master *pgxpool.Pool
	logger *slog.Logger
}

func NewClient(dsn string, opts ...Option) (*Client, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{master: pool, logger: slog.New(slog.DiscardHandler)}

	for _, opt := range opts {
		if err = opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) Master() *pgxpool.Pool {
	return c.master
}

func (c *Client) DB() *sql.DB {
	return stdlib.OpenDBFromPool(c.master)
}

func (c *Client) Ping(ctx context.Context) error {
	return c.master.Ping(ctx)
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	c.logger.Debug(query, logArgs(args)...)
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, query, args...)
	}
	return c.master.Exec(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	c.logger.Debug(query, logArgs(args)...)
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, query, args...)
	}
	return c.master.Query(ctx, query, args...)
}

func (c *Client) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	c.logger.Debug(query, logArgs(args)...)
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return c.master.QueryRow(ctx, query, args...)
}

func (c *Client) ScanQuery(ctx context.Context, dst any, query string, args ...any) error {
	c.logger.Debug(query, logArgs(args)...)
	rows, err := c.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() {
		rows.Close()
	}()
	return pgxscan.ScanAll(dst, rows)
}

func (c *Client) ScanQueryRow(ctx context.Context, dst any, query string, args ...any) error {
	c.logger.Debug(query, logArgs(args)...)
	rows, err := c.master.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return pgxscan.ScanOne(dst, rows)
}

func (c *Client) Close(_ context.Context) error {
	c.master.Close()
	return nil
}

func logArgs(args []any) []any {
	attr := make([]any, 0, len(args)*2) //nolint:mnd
	for i, arg := range args {
		k := fmt.Sprintf("$%d", i+1)
		attr = append(attr, slog.Any(k, arg))
	}
	return attr
}
