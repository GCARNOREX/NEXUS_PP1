package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const liveTTL = 10 * time.Second

// initRedis connects if REDIS_ADDR is set; without it the live cache is simply off.
func initRedis(ctx context.Context) error {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return nil
	}
	c := redis.NewClient(&redis.Options{Addr: addr})
	if err := c.Ping(ctx).Err(); err != nil {
		return err
	}
	rdb = c
	return nil
}

// cacheLive stores the latest snapshot for a host; the key expires if the host goes quiet.
func cacheLive(ctx context.Context, s Snapshot) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "live:"+s.Host, b, liveTTL).Err()
}

func liveSnapshots(ctx context.Context) ([]Snapshot, error) {
	var out []Snapshot
	iter := rdb.Scan(ctx, 0, "live:*", 100).Iterator()
	for iter.Next(ctx) {
		v, err := rdb.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue // expired between SCAN and GET
		}
		var s Snapshot
		if json.Unmarshal(v, &s) == nil {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Host < out[j].Host })
	return out, iter.Err()
}

func live(w http.ResponseWriter, r *http.Request) {
	if rdb == nil {
		http.Error(w, "redis not configured", http.StatusServiceUnavailable)
		return
	}
	snaps, err := liveSnapshots(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if snaps == nil {
		snaps = []Snapshot{}
	}
	writeJSON(w, snaps)
}
