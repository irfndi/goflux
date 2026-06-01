// Package telemetry provides opt-in anonymized telemetry for the goflux library.
// Users must explicitly call Enable() to activate reporting. No data is sent
// without explicit user consent.
package telemetry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	libVersion = "0.0.6"
	schemaVer  = 1
)

// Payload represents a single telemetry event.
type Payload struct {
	V           int               `json:"v"`
	Ts          int64             `json:"ts"`
	LibVersion  string            `json:"lib_version"`
	GoVersion   string            `json:"go_version"`
	OS          string            `json:"os"`
	Arch        string            `json:"arch"`
	Type        string            `json:"type"`
	Feature     string            `json:"feature,omitempty"`
	ErrorType   string            `json:"error_type,omitempty"`
	ErrorHash   string            `json:"error_hash,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
}

// Reporter handles telemetry reporting. A nil Reporter is safe to use
// and silently discards all reports.
type Reporter struct {
	enabled       bool
	endpoint      string
	token         string
	client        *http.Client
	buffer        chan Payload
	wg            sync.WaitGroup
	closeOnce     sync.Once
	closeCh       chan struct{}
	flushInterval time.Duration
	batchSize     int
}

var (
	defaultReporter *Reporter
	mu              sync.RWMutex
)

// Enable activates telemetry reporting to the given endpoint.
// The endpoint should be a URL like "https://goflux-telemetry.your-account.workers.dev/v1/telemetry".
// An optional bearer token can be provided for basic authentication.
func Enable(endpoint, token string) {
	mu.Lock()
	defer mu.Unlock()

	if defaultReporter != nil {
		defaultReporter.Close()
	}

	defaultReporter = &Reporter{
		enabled:       true,
		endpoint:      endpoint,
		token:         token,
		client:        &http.Client{Timeout: 10 * time.Second},
		buffer:        make(chan Payload, 100),
		closeCh:       make(chan struct{}),
		flushInterval: 30 * time.Second,
		batchSize:     10,
	}

	defaultReporter.wg.Add(1)
	go defaultReporter.flushLoop()
}

// Enabled reports whether telemetry is currently active.
func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return defaultReporter != nil && defaultReporter.enabled
}

// Disable stops telemetry reporting and flushes pending events.
func Disable() {
	mu.Lock()
	defer mu.Unlock()
	if defaultReporter != nil {
		defaultReporter.Close()
		defaultReporter = nil
	}
}

// ReportError reports an error occurrence. The actual error message is hashed
// to protect user privacy. Only the error type and a deterministic hash are sent.
func ReportError(err error, context map[string]string) {
	if err == nil {
		return
	}
	mu.RLock()
	r := defaultReporter
	mu.RUnlock()

	if r == nil || !r.enabled {
		return
	}

	msg := err.Error()
	hash := sha256.Sum256([]byte(msg))

	p := Payload{
		V:          schemaVer,
		Ts:         time.Now().UnixMilli(),
		LibVersion: libVersion,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Type:       "error",
		ErrorType:  fmt.Sprintf("%T", err),
		ErrorHash:  hex.EncodeToString(hash[:8]),
		Context:    context,
	}

	r.send(p)
}

// ReportUsage reports that a feature was used. This is useful for understanding
// which indicators, strategies, or utilities are most commonly used.
func ReportUsage(feature string, context map[string]string) {
	mu.RLock()
	r := defaultReporter
	mu.RUnlock()

	if r == nil || !r.enabled {
		return
	}

	p := Payload{
		V:          schemaVer,
		Ts:         time.Now().UnixMilli(),
		LibVersion: libVersion,
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		Type:       "usage",
		Feature:    feature,
		Context:    context,
	}

	r.send(p)
}

func (r *Reporter) send(p Payload) {
	select {
	case r.buffer <- p:
	default:
		// Buffer full; drop silently to avoid blocking user code.
	}
}

func (r *Reporter) flushLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()

	batch := make([]Payload, 0, r.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		for _, p := range batch {
			r.post(p)
		}
		batch = batch[:0]
	}

	for {
		select {
		case p, ok := <-r.buffer:
			if !ok {
				flush()
				return
			}
			batch = append(batch, p)
			if len(batch) >= r.batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-r.closeCh:
			flush()
			return
		}
	}
}

func (r *Reporter) post(p Payload) {
	body, err := json.Marshal(p)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", r.endpoint, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if r.token != "" {
		req.Header.Set("Authorization", "Bearer "+r.token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// Close shuts down the reporter, flushing pending events.
func (r *Reporter) Close() {
	r.closeOnce.Do(func() {
		close(r.closeCh)
		r.wg.Wait()
		close(r.buffer)
	})
}
