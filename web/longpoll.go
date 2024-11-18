package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// LongPollResponse handles the lifecycle of a long-polled HTTP response
type LongPollResponse struct {
	w    io.Writer
	f    http.Flusher
	Done chan struct{}
}

// WriteJSON writes the given payload to the client connection
func (s *LongPollResponse) WriteJSON(data any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.Write(encoded)
}

// Write writes the raw bytes to the client connection
func (s *LongPollResponse) Write(data []byte) error {
	_, err := s.w.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	s.f.Flush()
	return nil
}

// ErrStreamingUnsupported indicates the client connection does not support streaming
var ErrStreamingUnsupported = errors.New("streaming unsupported")

// NewLongPollResponse creates a new LongPollResponse for the given context and ResponseWriter
func NewLongPollResponse(context context.Context, w http.ResponseWriter) (*LongPollResponse, error) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return nil, ErrStreamingUnsupported
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	done := make(chan struct{})

	go func() {
		<-context.Done()
		close(done)
	}()

	return &LongPollResponse{
		w:    w,
		f:    f,
		Done: done,
	}, nil
}
