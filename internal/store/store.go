package store

import (
	"encoding/json"
	"sync"
)

type Store struct {
	mu       sync.Mutex
	Printers map[string]string
}

func NewStore() *Store {
	return &Store{
		Printers: make(map[string]string),
	}
}

func (s *Store) SetPrinter(id, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Printers[id] = data
}

func (s *Store) GetPrinters() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	printerData, _ := json.Marshal(s.Printers)
	return string(printerData)
}
