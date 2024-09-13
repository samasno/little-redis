package main

import (
	"os"
	"sync"
	"time"
)

type AppendOnlyFile struct {
	file *os.File
	mu   sync.Mutex
}

func NewAof(path string) (*AppendOnlyFile, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &AppendOnlyFile{
		file: f,
	}

	go func() {
		aof.mu.Lock()
		aof.file.Sync()
		aof.mu.Unlock()
		time.Sleep(time.Second)
	}()

	return aof, nil
}

func (a *AppendOnlyFile) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.file.Close()
}

func (a *AppendOnlyFile) Write(v Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, err := a.file.Write(v.Marshal())

	if err != nil {
		return err
	}

	return nil
}

func (a *AppendOnlyFile) Read(b []byte) (int, error) {
	return a.file.Read(b)
}
