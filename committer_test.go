package slogwit

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"syscall"
	"testing"
	"time"
)

const defaultTestQuickWitUrl = "http://localhost:7280"

func TestCommitterInit(t *testing.T) {
	const indexId = "testing_index1"
	if shouldSkip(t, indexId) {
		return
	}

	// should create index
	committer, err := NewCommitter(defaultTestQuickWitUrl, indexId)
	if err != nil {
		t.Fatal(err)
	}
	committer.Close()

	// should not fail on existing index
	committer, err = NewCommitter(defaultTestQuickWitUrl, indexId)
	if err != nil {
		t.Fatal(err)
	}
	committer.Close()

}

func TestCommitterWrite(t *testing.T) {
	const indexId = "testing_index2"
	if shouldSkip(t, indexId) {
		return
	}

	// should create index
	committer, err := NewCommitter(defaultTestQuickWitUrl, indexId)
	if err != nil {
		t.Fatal(err)
	}

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

	n, err := committer.Write(ndjsonFormatter(entries), len(entries))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(entries) {
		t.Fatal(io.ErrShortWrite)
	}
	committer.Close()

}

// Check if quickwit is up and Delete index
func shouldSkip(t *testing.T, indexId string) bool {
	deleteUrl, _ := url.JoinPath(defaultTestQuickWitUrl, apiPrefix, "indexes", indexId)
	req, err := http.NewRequest(http.MethodDelete, deleteUrl, nil)
	if err != nil {
		t.Fatal(err)
		return true
	}

	resp, err := http.DefaultClient.Do(req)
	if errors.Is(err, syscall.ECONNREFUSED) {
		t.SkipNow()
	}
	if resp != nil {
		resp.Body.Close()
	}

	return t.Skipped()
}
