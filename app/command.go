package main

import (
	"math"
	"strconv"
	"strings"
	"time"
)

type Command string

const (
	CommandPing Command = "PING"
	CommandEcho Command = "ECHO"
	CommandSet  Command = "SET"
	CommandGet  Command = "GET"
)

type SetOption string

const (
	SetOptionEX SetOption = "EX"
	SetOptionPX SetOption = "PX"
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
		if len(req.Array) != 3 && len(req.Array) != 5 {
			return RESPValue{Kind: RESPError, String: "ERR wrong number of arguments for 'set' command"}
		}
		key := req.Array[1].String
		value := req.Array[2].String
		entry := Entry{value: value}
		if len(req.Array) == 5 {
			option := SetOption(strings.ToUpper(req.Array[3].String))
			amount, err := strconv.ParseInt(req.Array[4].String, 10, 64)
			if err != nil {
				return RESPValue{Kind: RESPError, String: "ERR invalid expire time"}
			}
			if amount <= 0 {
				return RESPValue{Kind: RESPError, String: "ERR invalid expire time"}
			}

			var unit time.Duration
			switch option {
			case SetOptionEX:
				unit = time.Second
			case SetOptionPX:
				unit = time.Millisecond
			default:
				return RESPValue{Kind: RESPError, String: "ERR expected EX or PX option for 'set' command"}
			}

			// time.Duration is an int64 count of nanoseconds. Reject values that
			// would overflow when converting the supplied unit to nanoseconds.
			maxAmount := int64(time.Duration(math.MaxInt64) / unit)
			if amount > maxAmount {
				return RESPValue{Kind: RESPError, String: "ERR invalid expire time"}
			}
			ttl := time.Duration(amount) * unit
			entry.expiresAt = time.Now().Add(ttl)
		}
		store.Set(key, entry)
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
