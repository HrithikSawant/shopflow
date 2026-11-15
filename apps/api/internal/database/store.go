package database

import "github.com/jackc/pgx/v5/pgxpool"

// DB is the shared connection pool, set once at startup.
var DB *pgxpool.Pool
