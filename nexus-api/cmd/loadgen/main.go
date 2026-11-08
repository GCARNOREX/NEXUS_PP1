package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	const hosts, perHost = 20, 100
	var wg sync.WaitGroup
	var okCount, rejected atomic.Int64
	start := time.Now()
	for i := 0; i < hosts; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for n := 0; n < perHost; n++ {
				body := fmt.Sprintf(`{"host":"node-%d","timestamp":%d,"cpu_percent":%d.5,"mem":{"total_kb":1000,"free_kb":500,"available_kb":600},"processes":[]}`,
					id, time.Now().Unix(), n%100)
				resp, err := http.Post("http://localhost:8080/ingest", "application/json", bytes.NewBufferString(body))
				if err != nil {
					fmt.Println("error:", err)
					return
				}
				if resp.StatusCode == http.StatusAccepted {
					okCount.Add(1)
				} else {
					rejected.Add(1)
				}
				resp.Body.Close()
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("accepted %d, rejected %d in %v\n", okCount.Load(), rejected.Load(), time.Since(start))
}
