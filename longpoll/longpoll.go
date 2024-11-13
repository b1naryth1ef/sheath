package util

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type LongPollResponse struct {
	w    io.Writer
	f    http.Flusher
	Done chan struct{}
}

func (s *LongPollResponse) WriteJSON(data any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.Write(encoded)
}

func (s *LongPollResponse) Write(data []byte) error {
	_, err := s.w.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	s.f.Flush()
	return nil
}

var ErrStreamingUnsupported = errors.New("streaming unsupported")

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
