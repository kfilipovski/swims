package usas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const authURL = "https://securityapi.usaswimming.org/security/DataHubAuth/GetSisenseAuthToken"

type authRequest struct {
	SessionID  int64  `json:"sessionId"`
	DeviceID   int64  `json:"deviceId"`
	HostID     string `json:"hostId"`
	RequestURL string `json:"requestUrl"`
}

type authResponse struct {
	AccessToken string `json:"accessToken"`
}

func getToken(httpClient *http.Client) (string, error) {
	body := authRequest{
		SessionID:  24475633,
		DeviceID:   2449543657,
		HostID:     "MTI0ODY1OTU2Mw==",
		RequestURL: "/datahub/usas/individualsearch/times",
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling auth request: %w", err)
	}

	req, err := http.NewRequest("POST", authURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://data.usaswimming.org")
	req.Header.Set("Referer", "https://data.usaswimming.org/")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request returned status %d", resp.StatusCode)
	}

	var result authResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding auth response: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token in auth response")
	}
	return result.AccessToken, nil
}
