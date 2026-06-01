package telemetry

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryDisabledByDefault(t *testing.T) {
	assert.False(t, Enabled())
}

func TestTelemetryEnableDisable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	Enable(ts.URL+"/v1/telemetry", "")
	assert.True(t, Enabled())

	Disable()
	assert.False(t, Enabled())
}

func TestReportErrorWhenDisabled(t *testing.T) {
	// Should not panic when telemetry is disabled
	ReportError(errors.New("test error"), nil)
}

func TestReportUsageWhenDisabled(t *testing.T) {
	// Should not panic when telemetry is disabled
	ReportUsage("SMA", nil)
}

func setFastFlush() {
	mu.Lock()
	defer mu.Unlock()
	if defaultReporter != nil {
		defaultReporter.configMu.Lock()
		defaultReporter.flushInterval = 50 * time.Millisecond
		defaultReporter.batchSize = 1
		defaultReporter.configMu.Unlock()
	}
}

func TestReportError(t *testing.T) {
	received := make(chan Payload, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p Payload
		_ = json.Unmarshal(body, &p)
		received <- p
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	Enable(ts.URL, "")
	setFastFlush()
	defer Disable()

	ReportError(errors.New("something went wrong"), map[string]string{"indicator": "EMA"})

	select {
	case p := <-received:
		assert.Equal(t, "error", p.Type)
		assert.Equal(t, libVersion, p.LibVersion)
		assert.Equal(t, runtime.Version(), p.GoVersion)
		assert.Equal(t, runtime.GOOS, p.OS)
		assert.Equal(t, runtime.GOARCH, p.Arch)
		assert.Equal(t, "*errors.errorString", p.ErrorType)
		assert.NotEmpty(t, p.ErrorHash)
		assert.Equal(t, "EMA", p.Context["indicator"])
		assert.Equal(t, schemaVer, p.V)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for telemetry payload")
	}
}

func TestReportUsage(t *testing.T) {
	received := make(chan Payload, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p Payload
		_ = json.Unmarshal(body, &p)
		received <- p
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	Enable(ts.URL, "")
	setFastFlush()
	defer Disable()

	ReportUsage("BollingerBands", map[string]string{"window": "20"})

	select {
	case p := <-received:
		assert.Equal(t, "usage", p.Type)
		assert.Equal(t, "BollingerBands", p.Feature)
		assert.Equal(t, "20", p.Context["window"])
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for telemetry payload")
	}
}

func TestReportErrorNil(t *testing.T) {
	// Should not panic with nil error
	ReportError(nil, nil)
}

func TestTokenSentWhenConfigured(t *testing.T) {
	received := make(chan *http.Request, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	Enable(ts.URL, "secret-token-123")
	setFastFlush()
	defer Disable()

	ReportUsage("Test", nil)

	select {
	case r := <-received:
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer secret-token-123", auth)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for request")
	}
}

func TestReporterClose(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	Enable(ts.URL, "")
	Disable()

	// Close should be idempotent / safe after Disable
	mu.RLock()
	r := defaultReporter
	mu.RUnlock()
	require.Nil(t, r)
}
