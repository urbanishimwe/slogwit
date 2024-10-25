package slogwit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sync/atomic"
)

// A quickwit committer with retention period set to 30 days.
// error usually comes from initializing quickwit index.
func NewCommitter(quickwitUrl, indexId string) (Committer, error) {
	return newCommit(quickwitUrl, indexId, defaultRetentionPeriod)
}

// Same as NewCommitter but with provided retention.
// To disable retention set retentionPeriod to empty string.
func NewCommitterWithRetentionPeriod(quickwitUrl, indexId, retentionPeriod string) (Committer, error) {
	return newCommit(quickwitUrl, indexId, retentionPeriod)
}

const (
	apiPrefix      = "api/v1"
	ndjsonMimeType = "application/x-ndjson"
)

// A default implementation of committer. its simply an http client used to call quickwit REST API.
// It will create an index(if not exists) when creating a new commiter instance.
type commit struct {
	client    *http.Client
	ingestUrl string
	closed    atomic.Bool
}

func newCommit(quickWitUrl, indexId, retentionPeriod string) (Committer, error) {
	// trust Go team
	client := &http.Client{}

	err := initIndex(client, quickWitUrl, indexId, retentionPeriod)
	if err != nil {
		return nil, err
	}

	// format: host:port/api/v1/indexId/ingest
	ingestUrl, err := url.JoinPath(quickWitUrl, apiPrefix, indexId, "ingest")
	if err != nil {
		return nil, err
	}

	return &commit{client: client, ingestUrl: ingestUrl}, nil
}

func (c *commit) Write(ndjsonData []byte, recordCount int) (int, error) {
	if c.closed.Load() {
		return 0, io.ErrClosedPipe
	}

	resp, err := c.client.Post(c.ingestUrl, ndjsonMimeType, bytes.NewReader(ndjsonData))
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var respBody ingestResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return 0, err
	}

	if respBody.NumsDocs == recordCount {
		return recordCount, nil
	}

	return respBody.NumsDocs, errors.New(respBody.Message)
}

func (c *commit) Close() error {
	if c.closed.Load() {
		return nil
	}
	c.client.CloseIdleConnections()
	c.closed.Store(true)
	return nil
}

func initIndex(client *http.Client, quickwitUrl, indexId, retentionPeriod string) error {
	// First Check if index already exist
	exist, err := describeIndex(client, quickwitUrl, indexId)
	if err != nil {
		return err
	}

	if exist {
		// already exist
		return nil
	}

	return createIndex(client, quickwitUrl, indexId, retentionPeriod)
}

func describeIndex(client *http.Client, quickwitUrl, indexId string) (exist bool, err error) {
	// format: host:port/api/v1/indexes/indexId/describe
	describeIndexUrl, err := url.JoinPath(quickwitUrl, apiPrefix, "indexes", indexId, "describe")
	if err != nil {
		return false, err
	}
	resp, err := client.Get(describeIndexUrl)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	var respBody describeIndexResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return false, err
	}

	if respBody.IndexId == indexId {
		// index already exists
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if respBody.Message != "" {
		return false, errors.New(respBody.Message)
	}

	return false, fmt.Errorf("unexpected response from quickwit %d", resp.StatusCode)
}

func createIndex(client *http.Client, quickwitUrl, indexId, retentionPeriod string) error {
	//format: host:port/api/v1/indexes
	createIndexUrl, _ := url.JoinPath(quickwitUrl, apiPrefix, "indexes")
	reqBody := bytes.NewReader([]byte(entryIndexDefaultConfig(indexId, retentionPeriod)))
	resp, err := client.Post(createIndexUrl, mime.TypeByExtension(".json"), reqBody)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// read an error from body
	respBody := describeIndexResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return err
	}

	if respBody.Message != "" {
		return errors.New(respBody.Message)
	}

	return fmt.Errorf("unexpected response from quickwit %d", resp.StatusCode)
}

type ingestResponse struct {
	NumsDocs int `json:"num_docs_for_processing"`

	// This a part of a failed response. Its here to use this struct for both cases
	Message string `json:"message"`
}

// body returned by describe index api
type describeIndexResponse struct {
	// it returns many fields but we are interested in this
	IndexId string `json:"index_id"`

	// This a part of a failed response. Its here to use this struct for both cases
	Message string `json:"message"`
}
