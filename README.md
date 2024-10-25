# SLOGWIT [![Go Reference](https://pkg.go.dev/badge/github.com/urbanishimwe/slogwit.svg)](https://pkg.go.dev/github.com/urbanishimwe/slogwit)
Slog and quickwit.

It uses batching to reduce the number of push requests to Quickwit by aggregating multiple records in a single batch request.

```go
logger, err := slogwit.DefaultLogger(quickwitUrl, indexId)
if err != nil {
	return err
}

// You can start to use logger object
```

Logger is an object that implements the following method

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

