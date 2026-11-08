package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
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

const historyLen = 300

var (
	mu       sync.RWMutex
	latest   = map[string]Snapshot{}
	history  = map[string][]Snapshot{}
	lastSeen = map[string]time.Time{}
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
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
	mu.Lock()
	latest[s.Host] = s
	h := history[s.Host]
	if len(h) >= historyLen {
		copy(h, h[1:])
		h = h[:historyLen-1]
	}
	history[s.Host] = append(h, s)
	lastSeen[s.Host] = time.Now()
	mu.Unlock()
	w.WriteHeader(http.StatusAccepted)
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
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingest)
	mux.HandleFunc("GET /hosts", listHosts)
	mux.HandleFunc("GET /hosts/{host}", getHost)
	mux.HandleFunc("GET /hosts/{host}/history", getHistory)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
