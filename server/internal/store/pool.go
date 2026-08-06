package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// openPool builds a pgxpool with env-tunable MaxConns / lifetimes.
//
//	PG_POOL_MAX_CONNS       default 20
//	PG_POOL_MIN_CONNS       default 0
//	PG_POOL_MAX_CONN_LIFETIME_SEC  default 3600
//	PG_POOL_MAX_CONN_IDLE_SEC      default 300
func openPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	cfg.MaxConns = int32(envInt("PG_POOL_MAX_CONNS", 20))
	cfg.MinConns = int32(envInt("PG_POOL_MIN_CONNS", 0))
	if life := envInt("PG_POOL_MAX_CONN_LIFETIME_SEC", 3600); life > 0 {
		cfg.MaxConnLifetime = time.Duration(life) * time.Second
	}
	if idle := envInt("PG_POOL_MAX_CONN_IDLE_SEC", 300); idle > 0 {
		cfg.MaxConnIdleTime = time.Duration(idle) * time.Second
	}
	cfg.HealthCheckPeriod = 30 * time.Second
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func envInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		log.Printf("%s=%q invalid, using %d", key, raw, def)
		return def
	}
	return v
}
