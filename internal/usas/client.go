package usas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const apiURL = "https://usaswimming.sisense.com/api/datasources"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{}}
}

func (c *Client) query(datasource string, metadata []JaqlMetadata) (*JaqlResponse, error) {
	token, err := getToken(c.httpClient)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JAQL request returned status %d", resp.StatusCode)
	}

	var result JaqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JAQL response: %w", err)
	}
	return &result, nil
}
