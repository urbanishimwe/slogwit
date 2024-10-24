package slogwit

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// a standard logging interface
type Logger interface {
	io.WriteCloser
	Debug(message string, labels ...string) error
	Info(message string, labels ...string) error
	Notice(message string, labels ...string) error
	Warning(message string, labels ...string) error
	Error(message string, labels ...string) error
	Critical(message string, labels ...string) error
	Alert(message string, labels ...string) error
	Emergency(message string, labels ...string) error
}

// Batcher holds pending items waiting to be sent to Quickwit, and it's used
// to reduce the number of push requests to Quickwit aggregating multiple records
// in a single batch request.
type Batcher interface {
	io.Closer
	Write(Entry) error
	WithMaxBytes(uint64) Batcher
	WithQueueSize(uint64) Batcher
	WithCommitTimeout(time.Duration) Batcher
	// spin entries processor thread
	Run()
}

// an ndjson CommitWriter that also implements io.Closer
type Committer interface {
	CommitWriter
	io.Closer
}

// an interface for ndjson data writer(quickwit ingest)
type CommitWriter interface {
	Write(ndjsonData []byte, recordsCount int) (int, error)
}

type Entry struct {
	// Timestamp is the time of the entry. If zero, the current time is used.
	Timestamp time.Time `json:"timestamp"`

	// Severity is the entry's severity level.
	Severity Severity `json:"severity,omitempty"`

	// Payload is the actual log message.
	Payload string `json:"payload"`

	// Labels optionally specifies labels for logs.
	// may be formatted as pairs of key/value.
	Labels []string `json:"labels,omitempty"`
}

func (e Entry) String() string {
	return fmt.Sprintf(
		"{time=%s level=%s msg=%s labels=%s}",
		e.Timestamp, e.Severity, e.Payload, strings.Join(e.Labels, ","),
	)
}

type Severity string

const (
	// Debug means debug or trace information.
	Debug Severity = "DEBUG"
	// Info means routine information, such as ongoing status or performance.
	Info Severity = "INFO"
	// Notice means normal but significant events, such as start up, shut down, or configuration.
	Notice Severity = "NOTICE"
	// Warning means events that might cause problems.
	Warning Severity = "WARNING"
	// Error means events that are likely to cause problems.
	Error Severity = "ERROR"
	// Critical means events that cause more severe problems or brief outages.
	Critical Severity = "CRITICAL"
	// Alert means a person must take an action immediately.
	Alert Severity = "ALERT"
	// Emergency means one or more systems are unusable.
	Emergency Severity = "EMERGENCY"
)

const (
	// The default size of entries(in NDJSON) batcher can hold in memory before deciding to commit.
	DefaultBatchMaxBytes uint64 = 3 * 1024 * 1024 // 3 MB

	// Write will block if queue has more than this amount of entries(1 entry/ms).
	DefaultBatchEntriesQueueSize uint64 = 1000

	// After this time Batcher commit(flushes) entries to the writer in NDJSON format.
	// it is the same as a default value of quickwit "commit_timeout_secs"
	DefaultCommitTimeout = time.Minute
)
