package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

var db string
var port int

func main() {
	flag.StringVar(&db, "db", "", "file path to store key value pairs. will be loaded into memory on start")
	flag.IntVar(&port, "port", 8080, "port to accept tcp connections")
	flag.Parse()

	if db == "" {
		panic("db file is required")
	}

	addr := fmt.Sprintf(":%d", port)
	srv, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("failed to start server")
		panic(err.Error())
	}
	defer srv.Close()

	aof, err := NewAof(db)
	if err != nil {
		panic(err)
	}
	defer aof.Close()

	log.Println("reading backups into memory")
	loadBackup(aof)
	log.Println("done reading backups")

	backupRequests := func(v Value) {
		aof.Write(v)
	}

	for {
		conn, err := srv.Accept()
		if err == net.ErrClosed {
			log.Println("server closed")
			break
		}

		if err != nil {
			log.Println("server error" + err.Error())
			panic(err.Error())
		}

		go handleConnection(conn, backupRequests)
	}

	log.Println("Closing server ")
}

func handleConnection(conn net.Conn, options ...RequestOptions) {
	defer conn.Close()
	for {
		resp := NewResp(conn)
		request, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Println("error reading from client: ", err.Error())
			break
		}
		for _, opt := range options {
			opt(request)
		}

		response := HandleRequest(request)
		conn.Write(response.Marshal())
	}
}

func loadBackup(db io.Reader) error {
	for {
		resp := NewResp(db)
		request, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Println("error reading from client: ", err.Error())
			break
		}

		HandleRequest(request)
	}

	return nil
}

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
	NULL    = '~'
)

type Value struct {
	typ byte
	str string
	num int
	bul string
	arr []Value
}

func (v *Value) Marshal() []byte {
	switch v.typ {
	case STRING:
		return v.marshalString()
	case ERROR:
		return v.marshalError()
	case BULK:
		return v.marshalBulk()
	case ARRAY:
		return v.marshalArray()
	case 0:
		return v.marshalNull()
	default:
		return []byte{}
	}
}

func (v *Value) String() string {
	return string(v.Marshal())
}

func (v *Value) marshalString() []byte {
	b := []byte{}
	b = append(b, STRING)
	b = append(b, v.str...)
	b = append(b, '\r', '\n')
	return b
}

func (v *Value) marshalError() []byte {
	b := []byte{}
	b = append(b, ERROR)
	b = append(b, v.str...)
	b = append(b, '\r', '\n')
	return b
}

func (v *Value) marshalBulk() []byte {
	b := []byte{}
	b = append(b, BULK)
	b = append(b, strconv.Itoa(len(v.bul))...)
	b = append(b, '\r', '\n')
	b = append(b, v.bul...)
	b = append(b, '\r', '\n')
	return b
}

func (v *Value) marshalArray() []byte {
	l := len(v.arr)
	b := []byte{}
	b = append(b, ARRAY)
	b = append(b, strconv.Itoa(l)...)
	b = append(b, '\r', '\n')

	for i := 0; i < l; i++ {
		b = append(b, v.arr[i].Marshal()...)
	}

	return b
}

func (v *Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

type Resp struct {
	reader *bufio.Reader
}

func (r *Resp) Read() (Value, error) {
	t, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch t {
	case BULK:
		return r.readBulk()
	case ARRAY:
		return r.readArray()
	case STRING:
		return r.readString()
	default:
		log.Println("unkown type: " + string(t))
		return Value{}, nil
	}
}

func (r *Resp) readString() (Value, error) {
	v := Value{}
	v.typ = STRING

	l, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	b := make([]byte, l)

	_, err = r.reader.Read(b)
	if err != nil {
		return v, err
	}

	v.str = string(b)

	return v, nil
}

func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = ARRAY
	l, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.arr = []Value{}
	for i := 0; i < l; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.arr = append(v.arr, val)
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = BULK

	l, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	b := make([]byte, l)
	_, err = r.reader.Read(b)
	if err != nil {
		return v, err
	}

	v.bul = string(b)

	r.readLine()

	return v, nil
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return int(i64), n, nil
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

type RequestOptions func(v Value)
