package slogwit

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestBatchCommitTimeout(t *testing.T) {
	wg := new(sync.WaitGroup)
	nowTime := time.Now()
	entries := []Entry{
		{
			Timestamp: nowTime.Add(time.Minute),
			Payload:   "First Message",
		},
		{
			Timestamp: nowTime.Add(time.Minute * 2),
			Payload:   "Second Message",
		},
		{
			Timestamp: nowTime.Add(time.Minute * 3),
			Payload:   "Third Message",
		},
	}
	commiter := &dummyWriter{
		expect: ndjsonFormatter(entries),
		t:      t,
		wg:     wg,
	}

	wg.Add(1)

	batcher := NewBatcher(commiter).WithCommitTimeout(time.Second)
	go batcher.Run()

	for _, entry := range entries {
		batcher.Write(entry)
	}
	batcher.Close()

	wg.Wait()
}

func TestBatchMaxBytes(t *testing.T) {
	wg := new(sync.WaitGroup)
	nowTime := time.Now()
	entries := []Entry{
		{
			Timestamp: nowTime.Add(time.Minute),
			Payload:   "First Message",
		},
		{
			Timestamp: nowTime.Add(time.Minute * 2),
			Payload:   "Second Message",
		},
		{
			Timestamp: nowTime.Add(time.Minute * 3),
			Payload:   "Third Message",
		},
	}
	commiter := &dummyWriter{
		expect: ndjsonFormatter(entries),
		t:      t,
		wg:     wg,
	}

	const flushCount = 10 // flush 10 times
	wg.Add(flushCount)

	batcher := NewBatcher(commiter).WithMaxBytes(uint64(len(ndjsonFormatter(entries))))
	go batcher.Run()

	for i := 0; i < flushCount; i++ {
		for _, entry := range entries {
			batcher.Write(entry)
		}
	}
	batcher.Close()

	wg.Wait()
}

type dummyWriter struct {
	expect []byte
	t      *testing.T
	wg     *sync.WaitGroup
}

func (d *dummyWriter) Write(b []byte, n int) (int, error) {
	defer d.wg.Done()
	if !bytes.Equal(d.expect, b) {
		d.t.Fatalf("dummyWriter: wrong encoded data")
	}
	return len(b), nil
}

func ndjsonFormatter(entries []Entry) []byte {
	var entriesJson [][]byte
	for _, entry := range entries {
		entryJson, _ := json.Marshal(entry)
		entriesJson = append(entriesJson, entryJson)
	}
	return bytes.Join(entriesJson, []byte("\n"))
}
