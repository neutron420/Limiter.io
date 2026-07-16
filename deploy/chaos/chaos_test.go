package chaos

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestRedisOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					req, _ := http.NewRequest("GET", "http://localhost:8080/status", nil)
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						log.Printf("[%d] Request failed: %v", id, err)
						continue
					}
					resp.Body.Close()
					log.Printf("[%d] Status: %d", id, resp.StatusCode)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestKafkaOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Logf("Request %d failed: %v", i, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 500 {
			t.Logf("Request %d returned %d", i, resp.StatusCode)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func TestDatabaseOutage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for i := 0; i < 30; i++ {
		select {
		case <-sigCh:
			t.Log("Received signal, stopping test")
			return
		case <-ticker.C:
			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/v1/projects/test"), nil)
			resp, err := httpClient.Do(req)
			if err != nil {
				t.Logf("Request %d error: %v", i, err)
				continue
			}
			resp.Body.Close()
			t.Logf("Request %d status: %d", i, resp.StatusCode)
		}
	}
}

func TestMain(m *testing.M) {
	log.Println("Starting chaos tests - ensure services are running")
	code := m.Run()
	log.Println("Chaos tests completed")
	os.Exit(code)
}
