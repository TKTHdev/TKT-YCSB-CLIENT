// test-server is a minimal in-memory key-value store for benchmarking ycsb-client.
//
// Protocol (line-based TCP text):
//
//	Client → Server: "SET key value\n"
//	Server → Client: "OK\n"
//
//	Client → Server: "GET key\n"
//	Server → Client: "OK value\n"   (key found)
//	                 "ERR\n"         (key not found)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type KVStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func (s *KVStore) Set(key, value string) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	return v, ok
}

func handleConn(conn net.Conn, store *KVStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, " ", 3)
		switch parts[0] {
		case "SET":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERR")
				continue
			}
			store.Set(parts[1], parts[2])
			fmt.Fprintln(conn, "OK")
		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR")
				continue
			}
			if v, ok := store.Get(parts[1]); ok {
				fmt.Fprintf(conn, "OK %s\n", v)
			} else {
				fmt.Fprintln(conn, "ERR")
			}
		default:
			fmt.Fprintln(conn, "ERR")
		}
	}
}

func main() {
	addr := flag.String("addr", "localhost:7000", "listen address")
	flag.Parse()

	store := &KVStore{data: make(map[string]string)}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	log.Printf("test-server listening on %s", *addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConn(conn, store)
	}
}
