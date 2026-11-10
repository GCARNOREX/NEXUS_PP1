package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Mem struct {
	TotalKB     uint64 `json:"total_kb"`
	FreeKB      uint64 `json:"free_kb"`
	AvailableKB uint64 `json:"available_kb"`
}

type Proc struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	Threads    int64   `json:"threads"`
	RssKB      uint64  `json:"rss_kb"`
	CPUTicks   uint64  `json:"cpu_ticks"`
	CPUPercent float64 `json:"cpu_percent"`
}

type Snapshot struct {
	Host       string  `json:"host"`
	Timestamp  int64   `json:"timestamp"`
	CPUPercent float64 `json:"cpu_percent"`
	Mem        Mem     `json:"mem"`
	Processes  []Proc  `json:"processes"`
}

const (
	historyLen = 300
	queueSize  = 5000
	numWorkers = 8
)

var (
	mu       sync.RWMutex
	latest   = map[string]Snapshot{}
	history  = map[string][]Snapshot{}
	lastSeen = map[string]time.Time{}

	queue   = make(chan Snapshot, queueSize)
	dropped atomic.Uint64
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// store does the actual map updates (this used to live inside ingest).
func store(s Snapshot) {
	mu.Lock()
	defer mu.Unlock()
	latest[s.Host] = s
	h := history[s.Host]
	if len(h) >= historyLen {
		copy(h, h[1:])
		h = h[:historyLen-1]
	}
	history[s.Host] = append(h, s)
	lastSeen[s.Host] = time.Now()
}

func worker() {
	// WORKER_DELAY (for example 20ms) simulates slow work so you can watch the queue fill up
	delay, _ := time.ParseDuration(os.Getenv("WORKER_DELAY"))
	for s := range queue {
		if delay > 0 {
			time.Sleep(delay)
		}
		store(s)
		if pool != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := persist(ctx, s); err != nil {
				log.Printf("persist: %v", err)
			}
			cancel()
		}
	}
}

func ingest(w http.ResponseWriter, r *http.Request) {
	var s Snapshot
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if s.Host == "" {
		http.Error(w, "missing host", http.StatusBadRequest)
		return
	}
	select {
	case queue <- s:
		w.WriteHeader(http.StatusAccepted)
	default: // queue full: reject instead of blocking the handler
		dropped.Add(1)
		http.Error(w, "queue full", http.StatusServiceUnavailable)
	}
}

func stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"queue_depth": len(queue),
		"queue_size":  queueSize,
		"dropped":     dropped.Load(),
	})
}

type hostStatus struct {
	Host       string  `json:"host"`
	SecondsAgo float64 `json:"last_seen_seconds_ago"`
}

func listHosts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	out := make([]hostStatus, 0, len(lastSeen))
	for h, t := range lastSeen {
		out = append(out, hostStatus{h, time.Since(t).Seconds()})
	}
	mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Host < out[j].Host })
	writeJSON(w, out)
}

func getHost(w http.ResponseWriter, r *http.Request) {
	host := r.PathValue("host")
	mu.RLock()
	s, ok := latest[host]
	mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, s)
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	host := r.PathValue("host")
	mu.RLock()
	h, ok := history[host]
	cp := append([]Snapshot(nil), h...) // copy so the lock can be released before encoding
	mu.RUnlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, cp)
}

func main() {
	if err := initDB(context.Background()); err != nil {
		log.Fatalf("db: %v", err)
	}
	for i := 0; i < numWorkers; i++ {
		go worker()
	}
	startConsumer(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingest)
	mux.HandleFunc("GET /stats", stats)
	mux.HandleFunc("GET /hosts", listHosts)
	mux.HandleFunc("GET /hosts/{host}", getHost)
	mux.HandleFunc("GET /hosts/{host}/history", getHistory)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
