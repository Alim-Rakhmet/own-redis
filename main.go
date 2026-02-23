package main

import (
	"net"
	"own-redis/cmd"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Item represents a single stored value and its expiration time.
type Item struct {
	Value   string
	Expires int64 // Stored in Unix nanoseconds; 0 means no expiration
}

// Store handles the thread-safe, in-memory key-value map.
type Store struct {
	mu   sync.RWMutex
	data map[string]Item
}

// NewStore initializes a new key-value store.
func NewStore() *Store {
	return &Store{
		data: make(map[string]Item),
	}
}

// Set inserts or updates a key-value pair, with an optional expiration in milliseconds.
func (s *Store) Set(key, value string, px int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expires int64
	if px > 0 {
		expires = time.Now().UnixNano() + px*int64(time.Millisecond)
	}

	s.data[key] = Item{
		Value:   value,
		Expires: expires,
	}
}

// Get retrieves a value by key. It actively removes the key if it has expired.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	item, exists := s.data[key]
	s.mu.RUnlock()

	if !exists {
		return "", false
	}

	// Check if the item has an expiration and if the current time has surpassed it
	if item.Expires > 0 && time.Now().UnixNano() > item.Expires {
		s.mu.Lock()
		// Double-check locking to ensure another goroutine hasn't already modified it
		if it, ok := s.data[key]; ok && it.Expires == item.Expires {
			delete(s.data, key)
		}
		s.mu.Unlock()
		return "", false
	}

	return item.Value, true
}

// printUsage outputs the exact help text required by the assignment.

func main() {
	cmd.Run()
	// portFlag := flag.String("port", "8080")

	// // Resolve the UDP address
	// addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+port)
	// if err != nil {
	// 	fmt.Printf("Error resolving UDP address: %v\n", err)
	// 	os.Exit(1)
	// }

	// // Start listening for UDP packets
	// conn, err := net.ListenUDP("udp", addr)
	// if err != nil {
	// 	fmt.Printf("Error starting UDP server: %v\n", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	// store := NewStore()
	// buf := make([]byte, 2048) // Buffer to hold incoming UDP packets

	// // Infinite loop to read incoming packets
	// for {
	// 	n, clientAddr, err := conn.ReadFromUDP(buf)
	// 	if err != nil {
	// 		continue // Skip errors and keep listening
	// 	}

	// 	reqStr := string(buf[:n])
	// 	// Spawn a new goroutine to handle the request concurrently
	// 	go handleRequest(conn, clientAddr, reqStr, store)
	// }
}

// handleRequest parses the command and interacts with the Store.
func handleRequest(conn *net.UDPConn, clientAddr *net.UDPAddr, req string, store *Store) {
	// strings.Fields handles arbitrary whitespace (spaces, tabs, newlines) gracefully
	fields := strings.Fields(req)
	if len(fields) == 0 {
		return
	}

	command := strings.ToUpper(fields[0])
	var response string

	switch command {
	case "PING":
		response = "PONG\n"

	case "SET":
		if len(fields) < 3 {
			response = "(error) ERR wrong number of arguments for 'SET' command\n"
			break
		}

		key := fields[1]
		var val string
		var px int64

		// Check if the PX option was provided (it requires at least 5 fields: SET KEY VAL PX TIME)
		if len(fields) >= 5 && strings.ToUpper(fields[len(fields)-2]) == "PX" {
			parsedPx, err := strconv.ParseInt(fields[len(fields)-1], 10, 64)
			if err == nil && parsedPx > 0 {
				px = parsedPx
				// Join everything between the KEY and PX as the value
				val = strings.Join(fields[2:len(fields)-2], " ")
			} else {
				// If PX time is invalid, treat the trailing terms as part of the value
				val = strings.Join(fields[2:], " ")
			}
		} else {
			// Join all arguments after the KEY as the value (e.g., "bar baz")
			val = strings.Join(fields[2:], " ")
		}

		store.Set(key, val, px)
		response = "OK\n"

	case "GET":
		if len(fields) < 2 {
			response = "(error) ERR wrong number of arguments for 'GET' command\n"
			break
		}

		key := fields[1]
		val, exists := store.Get(key)
		if !exists {
			response = "(nil)\n"
		} else {
			response = val + "\n"
		}

	default:
		response = "(error) ERR unknown command\n"
	}

	// Send the response back to the client
	conn.WriteToUDP([]byte(response), clientAddr)
}
