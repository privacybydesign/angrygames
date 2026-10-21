package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	irmaRequestTimeout = 15 * time.Second
	maxIrmaResponse    = 1 << 20
	irmaSessionPath    = "/session"

	headerAuthorization = "Authorization"
	headerContentType   = "Content-Type"
	contentTypeJSON     = "application/json"
)

type irmaClient struct {
	serverURL string
	authToken string
	client    *http.Client
}

func newIrmaClient(serverURL, authToken string) *irmaClient {
	return &irmaClient{
		serverURL: serverURL,
		authToken: authToken,
		client:    &http.Client{Timeout: irmaRequestTimeout},
	}
}

// startSession posts the request to the IRMA server and returns the server's
// response body unchanged.
func (c *irmaClient) startSession(ctx context.Context, request disclosureRequest) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL+irmaSessionPath, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set(headerContentType, contentTypeJSON)
	if c.authToken != "" {
		req.Header.Set(headerAuthorization, c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IRMA server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxIrmaResponse))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, errors.New("IRMA server returned an empty response")
	}
	return body, nil
}
