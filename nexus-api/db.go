package main

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// initDB connects if DATABASE_URL is set; without it the API runs in-memory only.
func initDB(ctx context.Context) error {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil
	}
	p, err := pgxpool.New(ctx, url)
	if err != nil {
		return err
	}
	if err := p.Ping(ctx); err != nil {
		return err
	}
	pool = p
	return nil
}

// persist saves one snapshot and its processes in a single transaction.
func persist(ctx context.Context, s Snapshot) error {
	ts := time.Unix(s.Timestamp, 0)
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`INSERT INTO hosts (host, first_seen, last_seen) VALUES ($1, $2, $2)
		 ON CONFLICT (host) DO UPDATE SET last_seen = EXCLUDED.last_seen`,
		s.Host, ts); err != nil {
		return err
	}

	var id int64
	if err := tx.QueryRow(ctx,
		`INSERT INTO snapshots (host, ts, cpu_percent, mem_total_kb, mem_free_kb, mem_available_kb)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		s.Host, ts, s.CPUPercent,
		int64(s.Mem.TotalKB), int64(s.Mem.FreeKB), int64(s.Mem.AvailableKB)).Scan(&id); err != nil {
		return err
	}

	rows := make([][]any, len(s.Processes))
	for i, p := range s.Processes {
		rows[i] = []any{id, p.PID, p.Name, p.State, p.Threads,
			int64(p.RssKB), int64(p.CPUTicks), p.CPUPercent}
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"process_samples"},
		[]string{"snapshot_id", "pid", "name", "state", "threads", "rss_kb", "cpu_ticks", "cpu_percent"},
		pgx.CopyFromRows(rows)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
