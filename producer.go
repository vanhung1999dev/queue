package main

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Producer interface {
	Start() error
}

type HTTPProducer struct {
	listenAddr string
	server     *Server
	producech  chan<- Message
}

func NewHTTPProducer(listenAddr string, producech chan Message) *HTTPProducer {
	return &HTTPProducer{
		listenAddr: listenAddr,
		producech:  producech,
	}
}

func (p *HTTPProducer) Start() error {
	slog.Info("HTTP transport started", "port", p.listenAddr)
	return http.ListenAndServe(p.listenAddr, p)
}

func (p *HTTPProducer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	switch r.Method {
	case http.MethodPost:
		// POST /publish/<topic>
		if len(parts) != 2 || parts[0] != "publish" {
			http.Error(w, "invalid publish path (expected /publish/<topic>)", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		p.producech <- Message{
			Topic: parts[1],
			Data:  body,
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("message accepted\n"))

	case http.MethodGet:
		// GET /consume/<topic>/<offset>
		if len(parts) == 3 && parts[0] == "consume" {
			topic := parts[1]
			offsetStr := parts[2]

			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				http.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}

			store := p.server.getStoreForTopic(topic)
			data, err := store.Get(offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNoContent)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}

		http.Error(w, "invalid GET endpoint", http.StatusNotFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
