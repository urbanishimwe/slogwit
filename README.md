# SLOGWIT
Slog and quickwit.

It uses batching techniques to push periodically to quickwit by implementing the follwoing interfaces:

```go
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
```


```go
// an ndjson CommitWriter that also implements io.Closer
type Committer interface {
	CommitWriter
	io.Closer
}

// an interface for ndjson data writer(quickwit ingest)
type CommitWriter interface {
	Write(ndjsonData []byte, recordsCount int) (int, error)
}
```


```go
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
```

