package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/withoutforget/fshare/internal/config"
)

type key struct{}

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(cfg config.PostgresConfig) (*Postgres, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	slog.Info("postgres connected", "host", cfg.Host, "db", cfg.Database)

	return &Postgres{Pool: pool}, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}

func (p *Postgres) WithConn(ctx context.Context) (context.Context, error) {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return ctx, fmt.Errorf("postgres: acquire conn: %w", err)
	}
	return context.WithValue(ctx, key{}, conn), nil
}

func ConnFromCtx(ctx context.Context) *pgxpool.Conn {
	conn, ok := ctx.Value(key{}).(*pgxpool.Conn)
	if !ok {
		return nil
	}
	return conn
}
