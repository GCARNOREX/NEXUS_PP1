package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
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

var (
	mu     sync.RWMutex
	latest = map[string]Snapshot{}
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
	mu.Unlock()
	w.WriteHeader(http.StatusAccepted)
}

func listHosts(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	hosts := make([]string, 0, len(latest))
	for h := range latest {
		hosts = append(hosts, h)
	}
	mu.RUnlock()
	sort.Strings(hosts)
	writeJSON(w, hosts)
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /ingest", ingest)
	mux.HandleFunc("GET /hosts", listHosts)
	mux.HandleFunc("GET /hosts/{host}", getHost)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}