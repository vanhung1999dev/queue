package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type StoreProducerFunc func() Storer

type Storer interface {
	Push([]byte) (int, error)
	Get(int) ([]byte, error)
	Cleanup()
	StartRetentionCleaner(interval time.Duration, stop <-chan struct{})
}

type storedMessage struct {
	data      []byte
	timestamp time.Time
}

type MemoryStore struct {
	mu        sync.RWMutex
	data      []storedMessage
	retention time.Duration
}

func NewMemoryStore(retention time.Duration) *MemoryStore {
	return &MemoryStore{
		data:      make([]storedMessage, 0),
		retention: retention,
	}
}

func (s *MemoryStore) Push(b []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = append(s.data, storedMessage{
		data:      b,
		timestamp: time.Now(),
	})
	return len(s.data) - 1, nil
}

func (s *MemoryStore) Get(offset int) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be smaller than 0")
	}
	if len(s.data)-1 < offset {
		return nil, fmt.Errorf("offset (%d) too high", offset)
	}

	msg := s.data[offset]
	if time.Since(msg.timestamp) > s.retention {
		return nil, fmt.Errorf("message expired")
	}
	return msg.data, nil
}

func (s *MemoryStore) Cleanup() {
	slog.Info("running cleanup", "topic", "messages_before", len(s.data))

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.retention)
	idx := 0
	for i, msg := range s.data {
		if msg.timestamp.After(cutoff) {
			idx = i
			break
		}
		// if none are after cutoff, idx should be len(s.data)
		if i == len(s.data)-1 && !msg.timestamp.After(cutoff) {
			idx = len(s.data)
		}
	}
	if idx > 0 {
		s.data = s.data[idx:]
	}
}

func (s *MemoryStore) StartRetentionCleaner(interval time.Duration, stop <-chan struct{}) {
	slog.Info("started retention cleaner", "topic", "interval", interval)

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.Cleanup()
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
}
