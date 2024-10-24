package slogwit

import (
	"time"
)

// either committer or batcher is nil it returns nil object
func NewLogger(committer Committer, batcher Batcher) Logger {
	if committer == nil || batcher == nil {
		return nil
	}
	return &log{committer, batcher}
}

// default Logger implementation
type log struct {
	committer Committer
	batcher   Batcher
}

func (l *log) Write(data []byte) (int, error) {
	return len(data), l.batcher.Write(newEntry("", string(data), nil))
}

func (l *log) Close() error {
	l.batcher.Close()
	l.committer.Close()
	return nil
}

func (l *log) Debug(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Debug, message, labels))
}

func (l *log) Info(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Info, message, labels))
}
func (l *log) Notice(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Notice, message, labels))
}

func (l *log) Warning(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Warning, message, labels))
}

func (l *log) Error(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Error, message, labels))
}

func (l *log) Critical(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Critical, message, labels))
}

func (l *log) Alert(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Alert, message, labels))
}

func (l *log) Emergency(message string, labels ...string) error {
	return l.batcher.Write(newEntry(Emergency, message, labels))
}

func newEntry(level Severity, message string, labels []string) Entry {
	return Entry{
		Timestamp: time.Now(),
		Severity:  level,
		Payload:   message,
		Labels:    labels,
	}
}