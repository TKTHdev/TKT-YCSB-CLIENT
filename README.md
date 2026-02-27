# ycsb-client

A YCSB (Yahoo! Cloud Serving Benchmark) client that sends GET/SET workloads to a key-value server over a simple TCP text protocol.

## Protocol

The server must implement the following line-based text protocol:

```
# Write
Client → Server:  SET key value\n
Server → Client:  OK\n

# Read
Client → Server:  GET key\n
Server → Client:  OK value\n   (key found)
                  ERR\n         (key not found or error)
```

## Workloads

| Name   | Write ratio | Read ratio |
|--------|-------------|------------|
| ycsb-a | 50%         | 50%        |
| ycsb-b | 5%          | 95%        |
| ycsb-c | 0%          | 100%       |

- Keys: uniformly random from `k0` to `k{N-1}`
- Values: 100-byte random alphanumeric string
- Duration: 10 seconds

## Build

```bash
# Client
cd ycsb-client
go build -o ycsb-client .

# Test in-memory KV server
cd test-server
go build -o test-server .
```

## Usage

```bash
# Start the test server
./test-server --addr localhost:7000

# Run a benchmark
./ycsb-client \
  --addr     localhost:7000 \  # server address (repeatable)
  --workload ycsb-a \          # workload type
  --workers  4 \               # concurrent workers
  --keys     100 \             # number of distinct keys
  --csv      result.csv        # CSV output path (optional)
```

### Multiple servers

Specifying `--addr` multiple times distributes workers round-robin across servers.

```bash
./ycsb-client \
  --addr localhost:7000 \
  --addr localhost:7001 \
  --workers 8 \
  --workload ycsb-b
```

## Output

### Stdout

```
[ycsb-client] addrs=[localhost:7000] workload=ycsb-a workers=4 keys=100
Benchmark completed
Total ops: 120345
Throughput: 12034.50 ops/sec
Avg latency: 0.33 ms
RESULT:ycsb-a,4,100,12034.50,0.33
```

### CSV file

A header row is written automatically when the file is created. Subsequent runs append rows.

```csv
workload,workers,keys,throughput_ops_sec,avg_latency_ms
ycsb-a,4,100,12034.50,0.33
ycsb-b,4,100,15200.00,0.26
ycsb-c,4,100,18500.00,0.21
```

## File structure

```
ycsb-client/
  main.go       CLI entry point
  benchmark.go  Workload generation, measurement, CSV output
  client.go     TCP client (GET/SET)
  go.mod
  test-server/
    main.go     In-memory KV server for testing
    go.mod
```
