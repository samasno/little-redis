package main

import (
	"fmt"
	"strings"
	"sync"
)

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}
var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

var OK = Value{typ: STRING, str: "OK"}
var NIL = Value{typ: NULL}

func HandleRequest(v Value) Value {
	cmd := strings.ToUpper(v.arr[0].bul)
	hd, ok := Handlers[cmd]
	if !ok {
		return Value{typ: ERROR, str: "command not found"}
	}
	return hd(v.arr[1:])
}

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
	"HSET": hset,
	"HGET": hget,
}

func ping(args []Value) Value {
	return Value{typ: STRING, str: "PONG"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{typ: ERROR, str: fmt.Sprintf("ERROR: Expected 1 arg got %d", len(args))}
	}

	k := args[0].bul

	SETsMu.RLock()
	v, ok := SETs[k]
	SETsMu.RUnlock()
	if !ok {
		return NIL
	}

	return Value{typ: BULK, bul: v}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: ERROR, str: fmt.Sprintf("ERROR: Expected 2 args got %d", len(args))}
	}
	k, v := args[0].bul, args[1].bul

	SETsMu.Lock()
	SETs[k] = v
	SETsMu.Unlock()

	return Value{typ: STRING, str: "OK"}
}

func hset(args []Value) Value {
	if len(args) != 3 {
		return Value{typ: ERROR, str: fmt.Sprintf("ERROR: Expected 3 args got %d", len(args))}
	}

	h, k, v := args[0].bul, args[1].bul, args[2].bul

	HSETsMu.Lock()

	_, ok := HSETs[h]

	if !ok {
		HSETs[h] = map[string]string{}
	}
	HSETs[h][k] = v

	HSETsMu.Unlock()

	return OK
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{typ: ERROR, str: fmt.Sprintf("ERROR: Expected 2 args got %d", len(args))}
	}

	h, k := args[0].bul, args[1].bul

	HSETsMu.RLock()
	_, ok := HSETs[h]
	if !ok {
		return NIL
	}
	v, ok := HSETs[h][k]
	if !ok {
		return NIL
	}

	HSETsMu.RUnlock()

	return Value{typ: BULK, bul: v}

}
