package usas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const apiURL = "https://usaswimming.sisense.com/api/datasources"

type Client struct {
	httpClient *http.Client
	mu         sync.Mutex
	auth       *authState
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{}}
}

func (c *Client) query(datasource string, metadata []JaqlMetadata) (*JaqlResponse, error) {
	token, err := c.token(false)
	if err != nil {
		return nil, err
	}

	reqBody := JaqlRequest{
		Metadata:   metadata,
		Datasource: datasource,
		By:         "ComposeSDK",
		Count:      10000,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling JAQL request: %w", err)
	}

	url := fmt.Sprintf("%s/%s/jaql?trc=sdk-ui-1.11.0", apiURL, datasource)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("creating JAQL request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JAQL request: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		token, err = c.token(true)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest("POST", url, bytes.NewReader(b))
		if err != nil {
			return nil, fmt.Errorf("creating JAQL request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("JAQL request after auth refresh: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JAQL request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result JaqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JAQL response: %w", err)
	}
	return &result, nil
}

func (c *Client) token(forceRefresh bool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !forceRefresh && c.auth != nil && c.auth.Token != "" {
		return c.auth.Token, nil
	}

	state, err := refreshAuth(c.httpClient, c.auth)
	if err != nil {
		return "", err
	}
	c.auth = state
	return state.Token, nil
}
