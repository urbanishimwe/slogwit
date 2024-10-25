package slogwit

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// returns nil if commitPipe is not provided
func NewBatcher(commitPipe CommitWriter) Batcher {
	if commitPipe == nil {
		return nil
	}
	return &batch{
		maxBytes:      DefaultBatchMaxBytes,
		commitTimeout: DefaultCommitTimeout,
		entriesQueue:  make(chan *Entry, DefaultBatchEntriesQueueSize),
		commitPipe:    commitPipe,
	}
}

// the default batcher implementations. it expect commitPipe to always return a non-nil error.
// After Close, write to batch will return io.ErrClosedPipe but buffered entries will be processed before returning completely.
// Close will always return nil.
type batch struct {
	entriesNDJSON bytes.Buffer
	entriesQueue  chan *Entry
	entriesCount  uint64
	maxBytes      uint64
	commitTimeout time.Duration
	commitPipe    CommitWriter
	mux           sync.RWMutex
	closed        bool // to avoid writing on a closed channel
	runOnce       sync.Once
}

func (b *batch) WithMaxBytes(maxBytes uint64) Batcher {
	if maxBytes == 0 {
		maxBytes = DefaultBatchMaxBytes
	}
	b.maxBytes = maxBytes
	return b
}

func (b *batch) WithQueueSize(queueSize uint64) Batcher {
	if queueSize == 0 {
		queueSize = DefaultBatchEntriesQueueSize
	}
	b.entriesQueue = make(chan *Entry, queueSize)
	return b
}

func (b *batch) WithCommitTimeout(commitTimeout time.Duration) Batcher {
	if commitTimeout == 0 {
		commitTimeout = DefaultCommitTimeout
	}
	b.commitTimeout = commitTimeout
	return b
}

func (b *batch) Write(e Entry) error {
	// This is lock behaves as a read lock for b.closed and a write lock for b.entriesQueue
	b.mux.RLock()
	defer b.mux.RUnlock()
	if b.closed {
		return io.ErrClosedPipe
	}
	b.entriesQueue <- &e
	return nil
}

func (b *batch) Run() {
	b.runOnce.Do(func() {
		go b.runDequeue()
	})
}

func (b *batch) Close() error {
	b.mux.Lock()
	defer b.mux.Unlock()
	close(b.entriesQueue)
	b.closed = true
	return nil
}

func (b *batch) runDequeue() {
	defer b.flush()
	for {
		select {
		case entry, ok := <-b.entriesQueue:
			if !ok {
				// batch closed
				return
			}

			entryJson, _ := json.Marshal(entry)
			if b.entriesNDJSON.Len() > 0 {
				// add ndjson delimeter
				b.entriesNDJSON.WriteByte('\n')
			}
			b.entriesNDJSON.Write(entryJson)
			b.entriesCount++

			if b.entriesNDJSON.Len() >= int(b.maxBytes) {
				b.flush()
			}

		case <-time.After(b.commitTimeout):
			b.flush()
		}
	}
}

func (b *batch) flush() {
	defer b.entriesNDJSON.Reset()
	if b.entriesNDJSON.Len() <= 0 {
		return
	}
	b.commitPipe.Write(b.entriesNDJSON.Bytes(), int(b.entriesCount))
	b.entriesCount = 0
}
