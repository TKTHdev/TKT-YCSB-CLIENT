**English** | [日本語](README.ja.md)

# ycsb-client

A YCSB (Yahoo! Cloud Serving Benchmark) client that sends GET/SET workloads to a key-value server over a simple TCP text protocol.

## Server Requirements

This section defines what a server must implement to be benchmarked by ycsb-client.

### Connection model

- The server must accept TCP connections on a configured address.
- Each worker goroutine opens **one persistent TCP connection** at startup and reuses it for the entire 10-second benchmark. The server must **not** close the connection between requests.
- The client sends requests **one at a time per connection** (no pipelining): it sends a request, waits for the response, then sends the next request. The server therefore processes requests sequentially within a single connection.
- Multiple workers run concurrently, so the server must handle **N simultaneous connections** where N is the `--workers` count.

### Wire format

All messages are **UTF-8 plain text**, one message per line, terminated by `\n` (LF only, not CRLF).

**SET request** (write):

```
SET <key> <value>\n
```

- `<key>`: alphanumeric string, no spaces (e.g. `k0`, `k42`)
- `<value>`: 100-byte alphanumeric string, no spaces

The server must store the key-value pair and reply **exactly**:

```
OK\n
```

**GET request** (read):

```
GET <key>\n
```

The server must reply **exactly one of**:

```
OK <value>\n    ← key exists: "OK" + one space + the stored value
ERR\n           ← key does not exist, or any error
```

> **Important:** Only requests that receive `OK` are counted toward throughput and latency. `ERR` responses are silently skipped. For YCSB-C (read-only workload), the store must be pre-populated before the benchmark, otherwise all GETs will return `ERR` and throughput will be zero.

### Summary table

| Scenario | Client sends | Server must reply |
|---|---|---|
| Write | `SET k value\n` | `OK\n` |
| Read (hit) | `GET k\n` | `OK value\n` |
| Read (miss) | `GET k\n` | `ERR\n` |
| Any error | — | `ERR\n` |

### Reference implementation

`test-server/` contains a minimal single-node in-memory KV server that satisfies this protocol. Use it as a reference when implementing the server side in your own distributed system.

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
