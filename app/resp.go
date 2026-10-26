package main

import (
	"fmt"
	"io"
	"strconv"
)

type RESPKind byte

const (
	RESPSimpleString RESPKind = '+'
	RESPError        RESPKind = '-'
	RESPInteger      RESPKind = ':'
	RESPBulkString   RESPKind = '$'
	RESPArray        RESPKind = '*'
)

type RESPValue struct {
	Kind    RESPKind
	String  string
	Integer int64
	Array   []RESPValue
}

func readRESPLine(input []byte) (line []byte, consumed int, err error) {
	for i, ch := range input {
		switch ch {
		case '\r':
			if i+1 >= len(input) {
				return nil, 0, io.ErrUnexpectedEOF
			}
			if input[i+1] != '\n' {
				return nil, 0, fmt.Errorf("expected LF after CR at byte %d", i+1)
			}
			return input[:i], i + 2, nil
		case '\n':
			return nil, 0, fmt.Errorf("unexpected LF at byte %d", i)
		}
	}

	return nil, 0, io.ErrUnexpectedEOF
}

func readRESPLength(input []byte) (length int, consumed int, err error) {
	header, consumed, err := readRESPLine(input)
	if err != nil {
		return 0, 0, err
	}

	length, err = strconv.Atoi(string(header))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid length %q: %w", header, err)
	}
	if length < 0 {
		return 0, 0, fmt.Errorf("negative length %d is not supported", length)
	}

	return length, consumed, nil
}

func decodeRESP(msg []byte) (value RESPValue, consumed int, err error) {
	if len(msg) == 0 {
		return RESPValue{}, 0, io.ErrUnexpectedEOF
	}

	prefix := msg[0]

	switch prefix {
	case '*':
		value, n, err := decodeArray(msg[1:])
		if err != nil {
			return RESPValue{}, 0, err
		}
		return value, n + 1, err
	case '$':
		value, n, err := decodeBulkString(msg[1:])
		if err != nil {
			return RESPValue{}, 0, err
		}
		return value, n + 1, err
	case '+':
		value, n, err := decodeSimpleString(msg[1:])
		if err != nil {
			return RESPValue{}, 0, err
		}
		return value, n + 1, err
	case '-':
		return RESPValue{}, 0, fmt.Errorf("RESP error values are not implemented")
	case ':':
		return RESPValue{}, 0, fmt.Errorf("RESP integer values are not implemented")
	default:
		return RESPValue{}, 0, fmt.Errorf("unsupported RESP prefix %q", prefix)
	}
}

func decodeArray(input []byte) (value RESPValue, consumed int, err error) {
	count, offset, err := readRESPLength(input)
	if err != nil {
		return RESPValue{}, 0, err
	}

	items := make([]RESPValue, 0)

	for range count {
		v, n, err := decodeRESP(input[offset:])
		if err != nil {
			return RESPValue{}, 0, err
		}

		items = append(items, v)
		offset += n
	}

	return RESPValue{Kind: RESPArray, Array: items}, offset, nil
}

func decodeSimpleString(input []byte) (value RESPValue, consumed int, err error) {
	line, consumed, err := readRESPLine(input)
	if err != nil {
		return RESPValue{}, 0, err
	}

	return RESPValue{
		Kind:   RESPSimpleString,
		String: string(line),
	}, consumed, nil
}

func decodeBulkString(input []byte) (value RESPValue, consumed int, err error) {
	length, bodyStart, err := readRESPLength(input)
	if err != nil {
		return RESPValue{}, 0, err
	}

	bytesAfterHeader := len(input) - bodyStart
	if bytesAfterHeader < 2 || length > bytesAfterHeader-2 {
		return RESPValue{}, 0, io.ErrUnexpectedEOF
	}
	bodyEnd := bodyStart + length
	if input[bodyEnd] != '\r' || input[bodyEnd+1] != '\n' {
		return RESPValue{}, 0, fmt.Errorf("expected CRLF after bulk string at byte %d", bodyEnd)
	}

	value = RESPValue{
		Kind:   RESPBulkString,
		String: string(input[bodyStart:bodyEnd]),
	}

	return value, bodyEnd + 2, nil
}

func encodeRESP(value RESPValue) ([]byte, error) {
	switch value.Kind {
	case RESPSimpleString, RESPError:
		return fmt.Appendf(nil, "%c%s\r\n", value.Kind, value.String), nil
	case RESPInteger:
		return fmt.Appendf(nil, ":%d\r\n", value.Integer), nil
	case RESPBulkString:
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(value.String), value.String), nil
	case RESPArray:
		encoded := fmt.Appendf(nil, "*%d\r\n", len(value.Array))
		for _, item := range value.Array {
			itemBytes, err := encodeRESP(item)
			if err != nil {
				return nil, err
			}
			encoded = append(encoded, itemBytes...)
		}
		return encoded, nil
	default:
		return nil, fmt.Errorf("unsupported RESP kind %q", byte(value.Kind))
	}
}
