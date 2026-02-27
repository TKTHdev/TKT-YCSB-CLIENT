package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// TCPClient sends GET/SET commands to a key-value server over a persistent TCP connection.
//
// Protocol (line-based text):
//
//	Client → Server: "SET key value\n"
//	Server → Client: "OK\n"
//
//	Client → Server: "GET key\n"
//	Server → Client: "OK value\n"   (key found)
//	                 "ERR\n"         (key not found or error)
type TCPClient struct {
	conn   net.Conn
	reader *bufio.Reader
	debug  bool
	addr   string
}

func NewTCPClient(addr string, debug bool) *TCPClient {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to %s: %v", addr, err))
	}
	return &TCPClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		debug:  debug,
		addr:   addr,
	}
}

func (c *TCPClient) Put(key, value string) bool {
	if _, err := fmt.Fprintf(c.conn, "SET %s %s\n", key, value); err != nil {
		c.log("PUT write error: %v", err)
		return false
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.log("PUT read error: %v", err)
		return false
	}
	return strings.TrimSpace(line) == "OK"
}

func (c *TCPClient) Get(key string) (string, bool) {
	if _, err := fmt.Fprintf(c.conn, "GET %s\n", key); err != nil {
		c.log("GET write error: %v", err)
		return "", false
	}
	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.log("GET read error: %v", err)
		return "", false
	}
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "OK") {
		return strings.TrimPrefix(line, "OK "), true
	}
	return "", false
}

func (c *TCPClient) Close() {
	c.conn.Close()
}

func (c *TCPClient) log(format string, args ...interface{}) {
	if c.debug {
		fmt.Printf("[TCPClient %s] %s\n", c.addr, fmt.Sprintf(format, args...))
	}
}
