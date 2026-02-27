package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	experimentDuration = 10 * time.Second
	valueSize          = 100
)

// Workload represents a YCSB workload type as the write percentage.
type Workload int

const (
	WorkloadA Workload = 50 // YCSB-A: 50% writes, 50% reads
	WorkloadB Workload = 5  // YCSB-B:  5% writes, 95% reads
	WorkloadC Workload = 0  // YCSB-C:  0% writes, 100% reads
)

// ParseWorkload converts "ycsb-a", "ycsb-b", "ycsb-c" to a Workload. Defaults to WorkloadA.
func ParseWorkload(s string) Workload {
	switch s {
	case "ycsb-b":
		return WorkloadB
	case "ycsb-c":
		return WorkloadC
	default:
		return WorkloadA
	}
}

func (w Workload) String() string {
	switch w {
	case WorkloadA:
		return "ycsb-a"
	case WorkloadB:
		return "ycsb-b"
	case WorkloadC:
		return "ycsb-c"
	default:
		return fmt.Sprintf("unknown(%d)", int(w))
	}
}

// Config holds all benchmark parameters.
type Config struct {
	// Addrs is the list of server addresses. Workers are distributed round-robin.
	Addrs    []string
	Workers  int
	NumKeys  int
	Workload Workload
	// CSVPath is the path to the output CSV file. If empty, results are only printed to stdout.
	CSVPath string
	Debug   bool
}

type workerResult struct {
	count    int
	duration time.Duration
}

// Run executes the YCSB benchmark. Each worker gets its own persistent TCP connection.
// Results are printed to stdout and optionally appended to a CSV file.
func Run(cfg Config) {
	fmt.Printf("[ycsb-client] addrs=%v workload=%s workers=%d keys=%d\n",
		cfg.Addrs, cfg.Workload, cfg.Workers, cfg.NumKeys)

	ctx, cancel := context.WithTimeout(context.Background(), experimentDuration)
	defer cancel()

	resultCh := make(chan workerResult, cfg.Workers)
	var wg sync.WaitGroup

	for i := 0; i < cfg.Workers; i++ {
		addr := cfg.Addrs[i%len(cfg.Addrs)]
		client := NewTCPClient(addr, cfg.Debug)
		wg.Add(1)
		go func(c *TCPClient) {
			defer wg.Done()
			defer c.Close()
			resultCh <- runWorker(ctx, c, cfg)
		}(client)
	}

	wg.Wait()
	close(resultCh)

	var totalCount int
	var totalDuration time.Duration
	for res := range resultCh {
		totalCount += res.count
		totalDuration += res.duration
	}

	throughput := float64(totalCount) / experimentDuration.Seconds()
	avgLatency := float64(0)
	if totalCount > 0 {
		avgLatency = float64(totalDuration.Milliseconds()) / float64(totalCount)
	}

	fmt.Printf("Benchmark completed\n")
	fmt.Printf("Total ops: %d\n", totalCount)
	fmt.Printf("Throughput: %.2f ops/sec\n", throughput)
	fmt.Printf("Avg latency: %.2f ms\n", avgLatency)
	fmt.Printf("RESULT:%s,%d,%d,%.2f,%.2f\n", cfg.Workload, cfg.Workers, cfg.NumKeys, throughput, avgLatency)

	if cfg.CSVPath != "" {
		if err := appendCSV(cfg, throughput, avgLatency); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write CSV: %v\n", err)
		}
	}
}

func runWorker(ctx context.Context, client *TCPClient, cfg Config) workerResult {
	res := workerResult{}
	writeRatio := int(cfg.Workload)

	for {
		select {
		case <-ctx.Done():
			return res
		default:
		}

		key := fmt.Sprintf("k%d", rand.Intn(cfg.NumKeys))
		value := randomValue(valueSize)

		start := time.Now()
		var ok bool

		if rand.Intn(100) < writeRatio {
			ok = client.Put(key, value)
		} else {
			_, ok = client.Get(key)
		}

		if ok {
			res.count++
			res.duration += time.Since(start)
		}
	}
}

func randomValue(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

var csvHeader = []string{"workload", "workers", "keys", "throughput_ops_sec", "avg_latency_ms"}

func appendCSV(cfg Config, throughput, avgLatency float64) error {
	needHeader := false
	if _, err := os.Stat(cfg.CSVPath); os.IsNotExist(err) {
		needHeader = true
	}

	f, err := os.OpenFile(cfg.CSVPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if needHeader {
		if err := w.Write(csvHeader); err != nil {
			return err
		}
	}
	row := []string{
		cfg.Workload.String(),
		fmt.Sprintf("%d", cfg.Workers),
		fmt.Sprintf("%d", cfg.NumKeys),
		fmt.Sprintf("%.2f", throughput),
		fmt.Sprintf("%.2f", avgLatency),
	}
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}
