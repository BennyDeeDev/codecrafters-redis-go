package main

import "strings"

type Command string

const (
	CommandPing Command = "PING"
	CommandEcho Command = "ECHO"
	CommandSet  Command = "SET"
	CommandGet  Command = "GET"
)

func handleCommand(req RESPValue, store *Store) RESPValue {
	if req.Kind != RESPArray || len(req.Array) == 0 {
		return RESPValue{Kind: RESPError, String: "ERR invalid command"}
	}

	command := Command(strings.ToUpper(req.Array[0].String))

	switch command {
	case CommandPing:
		switch len(req.Array) {
		case 1:
			return RESPValue{Kind: RESPSimpleString, String: "PONG"}
		case 2:
			return RESPValue{Kind: RESPBulkString, String: req.Array[1].String}
		default:
			return RESPValue{Kind: RESPError, String: "ERR wrong number of arguments for 'ping' command"}
		}
	case CommandEcho:
		if len(req.Array) != 2 {
			return RESPValue{Kind: RESPError, String: "ERR wrong number of arguments for 'echo' command"}
		}
		return RESPValue{Kind: RESPBulkString, String: req.Array[1].String}
	case CommandSet:
		if len(req.Array) != 3 {
			return RESPValue{Kind: RESPError, String: "ERR wrong number of arguments for 'set' command"}
		}
		store.Set(req.Array[1].String, req.Array[2].String)
		return RESPValue{Kind: RESPSimpleString, String: "OK"}
	case CommandGet:
		if len(req.Array) != 2 {
			return RESPValue{Kind: RESPError, String: "ERR wrong number of arguments for 'get' command"}
		}
		value, ok := store.Get(req.Array[1].String)
		if !ok {
			return RESPValue{Kind: RESPBulkString, Null: true}
		}
		return RESPValue{Kind: RESPBulkString, String: value}

	default:
		return RESPValue{Kind: RESPError, String: "ERR unknown command"}
	}
}
