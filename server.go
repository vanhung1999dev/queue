package main

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type Message struct {
	Topic string
	Data  []byte
}

type Config struct {
	ListenAddr        string
	StoreProducerFunc StoreProducerFunc
	CleanupInterval   time.Duration
}

type Server struct {
	*Config
	mu           sync.RWMutex
	topics       map[string]Storer
	cleanerStops map[string]chan struct{}
	consumers    []Consumer
	producers    []Producer
	producech    chan Message
	quitch       chan struct{}
}

func NewServer(cfg *Config) (*Server, error) {
	producech := make(chan Message)

	s := &Server{
		Config:       cfg,
		topics:       make(map[string]Storer),
		cleanerStops: make(map[string]chan struct{}),
		quitch:       make(chan struct{}),
		producech:    producech,
	}

	httpProducer := NewHTTPProducer(cfg.ListenAddr, producech)
	httpProducer.server = s // Required!

	s.producers = []Producer{httpProducer}
	return s, nil
}

func (s *Server) Start() {
	// for _, consumer := range s.consumers {
	// 	if err := consumer.Start(); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }
	for _, producer := range s.producers {
		go func(p Producer) {
			if err := p.Start(); err != nil {
				fmt.Println(err)
			}
		}(producer)
	}
	s.loop()
}

func (s *Server) loop() {
	for {
		select {
		case <-s.quitch:
			return
		case msg := <-s.producech:
			offset, err := s.publish(msg)
			if err != nil {
				slog.Error("failed to publish", err)
			} else {
				slog.Info("produced message", "offset", offset)
			}
		}
	}
}

func (s *Server) publish(msg Message) (int, error) {
	store := s.getStoreForTopic(msg.Topic)
	return store.Push(msg.Data)
}

func (s *Server) getStoreForTopic(topic string) Storer {
	s.mu.RLock()
	store, ok := s.topics[topic]
	s.mu.RUnlock()
	if ok {
		return store
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after locking
	store, ok = s.topics[topic]
	if ok {
		return store
	}

	store = s.StoreProducerFunc()

	stop := make(chan struct{})
	store.StartRetentionCleaner(s.CleanupInterval, stop)

	s.topics[topic] = store
	s.cleanerStops[topic] = stop

	slog.Info("created new topic", "topic", topic)
	slog.Info("started retention cleaner", "topic", topic, "interval", s.CleanupInterval.String())
	return store
}
