package server

import (
	"log/slog"
	"net"
	"own-redis/internal"
	"own-redis/internal/store"
	"strconv"
	"strings"
)

func Start(port string, store *store.Store) error {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:"+port)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	slog.Info("Started UDP sever", "port", port)

	buffer := make([]byte, 2048)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			slog.Error("Failed to read from UDP", "error", err)
			continue
		}

		request := buffer[:n]

		go handleRequest(conn, clientAddr, string(request), store)
	}
}

func handleRequest(conn *net.UDPConn, addr *net.UDPAddr, request string, store *store.Store) {
	fields := strings.Fields(request)

	if len(fields) == 0 {
		return
	}

	command := strings.ToUpper(fields[0])
	var response string

	switch command {
	case "PING":
		response = "PONG"

	case "GET":
		if len(fields) == 1 || len(fields) > 2 {
			response = internal.ErrWrongNumOfArgs.Error()
			break
		}

		value, result := store.Get(fields[1])
		if result {
			response = value
		} else {
			response = "(nil)"
		}

	case "SET":
		if len(fields) < 3 {
			response = internal.ErrWrongNumOfArgs.Error()
			break
		}

		var input []string
		var timeOut int64
		for i, field := range fields {
			if i == 0 {
				continue
			}

			if strings.ToLower(field) == "px" && i == len(fields)-2 {
				px, err := strconv.Atoi(fields[len(fields)-1])
				if err == nil {
					timeOut = int64(px)
					break
				}
			}

			input = append(input, field)
		}

		if response == "" {
			response = "OK"

			key := input[0]
			value := strings.Join(input[1:], " ")
			store.Set(key, value, timeOut)
		}

	default:
		response = internal.ErrUnknownCommand.Error()
	}

	response += string('\n')
	conn.WriteToUDP([]byte(response), addr)
}
