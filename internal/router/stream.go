package router

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SSEWriter writes Server-Sent Events to an http.ResponseWriter.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter creates a new SSEWriter.
func NewSSEWriter(w http.ResponseWriter) *SSEWriter {
	flusher, _ := w.(http.Flusher)
	return &SSEWriter{w: w, flusher: flusher}
}

// WriteEvent writes a single SSE event.
func (sw *SSEWriter) WriteEvent(event string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal SSE data: %w", err)
	}

	if event != "" {
		fmt.Fprintf(sw.w, "event: %s\n", event)
	}
	fmt.Fprintf(sw.w, "data: %s\n\n", jsonData)

	if sw.flusher != nil {
		sw.flusher.Flush()
	}
	return nil
}

// WriteRawEvent writes a raw SSE event with pre-encoded data.
func (sw *SSEWriter) WriteRawEvent(event string, data []byte) {
	if event != "" {
		fmt.Fprintf(sw.w, "event: %s\n", event)
	}
	fmt.Fprintf(sw.w, "data: %s\n\n", data)

	if sw.flusher != nil {
		sw.flusher.Flush()
	}
}

// SSEReader reads Server-Sent Events from a response body.
type SSEReader struct {
	scanner *bufio.Scanner
}

// NewSSEReader creates a new SSEReader from a response body.
func NewSSEReader(body io.Reader) *SSEReader {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	return &SSEReader{scanner: scanner}
}

// ReadEvent reads the next SSE event. Returns event type and data.
// Returns io.EOF when the stream ends.
func (r *SSEReader) ReadEvent() (event string, data string, err error) {
	var eventType string
	var dataLines []string

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line signals end of event
		if line == "" {
			if len(dataLines) > 0 {
				return eventType, strings.Join(dataLines, "\n"), nil
			}
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			dataLines = append(dataLines, strings.TrimPrefix(line, "data: "))
		} else if line == "data:" {
			dataLines = append(dataLines, "")
		}
		// Ignore comments (lines starting with :) and other fields
	}

	if err := r.scanner.Err(); err != nil {
		return "", "", err
	}

	// If we have remaining data, return it
	if len(dataLines) > 0 {
		return eventType, strings.Join(dataLines, "\n"), nil
	}

	return "", "", io.EOF
}

// SetSSEHeaders sets the appropriate headers for SSE responses.
func SetSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}
