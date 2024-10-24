package slogwit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sync/atomic"
)

func NewCommitter(quickwitUrl, indexId string) (Committer, error) {
	return newCommit(quickwitUrl, indexId)
}

const (
	apiPrefix      = "api/v1"
	ndjsonMimeType = "application/x-ndjson"
)

// A default implementation of committer. its simply an http client used to call quickwit REST API.
// Committer may initialize index when creating new commiter instance
type commit struct {
	client    *http.Client
	ingestUrl string
	closed    atomic.Bool
}

func newCommit(quickWitUrl string, indexId string) (Committer, error) {
	// trust Go team
	client := http.DefaultClient

	err := initIndex(client, quickWitUrl, indexId)
	if err != nil {
		return nil, err
	}

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

func initIndex(client *http.Client, quickwitUrl, indexId string) error {
	// First Check if index already exist
	describeIndexUrl, err := url.JoinPath(quickwitUrl, apiPrefix, "indexes", indexId, "describe")
	if err != nil {
		return err
	}
	resp, err := client.Get(describeIndexUrl)
	if err != nil {
		return nil
	}

	var respBody describeIndexResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if respBody.IndexId == indexId {
		// index already exists
		return nil
	}

	if resp.StatusCode != http.StatusNotFound {
		return errors.New(respBody.Message)
	}

	// Now create an index
	createIndexUrl, _ := url.JoinPath(quickwitUrl, apiPrefix, "indexes")
	reqBody := bytes.NewReader([]byte(entryIndexDefaultConfig(indexId)))
	resp, err = client.Post(createIndexUrl, mime.TypeByExtension(".json"), reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// read an error from body
	respBody = describeIndexResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return err
	}

	return errors.New(respBody.Message)
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
