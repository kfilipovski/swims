package usas

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

const (
	authURL         = "https://securityapi.usaswimming.org/security/DataHubAuth/GetSisenseAuthToken"
	securityInfoURL = "https://securityapi.usaswimming.org/security/Auth/GetSecurityInfoForToken"
	requestURL      = "/datahub/usas/individualsearch/times"
)

var publicIPURLs = []string{
	"https://ipv4.icanhazip.com/",
	"https://api.ipify.org/",
}

type authRequest struct {
	SessionID  int64  `json:"sessionId"`
	DeviceID   string `json:"deviceId"`
	HostID     string `json:"hostId"`
	RequestURL string `json:"requestUrl"`
}

type securityInfoRequest struct {
	AccessToken string   `json:"access_token"`
	Toxonomies  []string `json:"toxonomies"`
	Scope       string   `json:"scope"`
	UIProject   string   `json:"uIProjectName"`
	BustCache   bool     `json:"bustCache"`
	AppName     string   `json:"appName"`
	DeviceID    string   `json:"deviceId"`
	HostID      string   `json:"hostId"`
}

type authResponse struct {
	AccessToken string `json:"accessToken"`
}

type securityInfoResponse struct {
	RequestID string `json:"requestId"`
}

type authState struct {
	Token    string
	DeviceID string
	HostID   string
}

func refreshAuth(httpClient *http.Client, state *authState) (*authState, error) {
	deviceID, err := getDeviceID(state)
	if err != nil {
		return nil, err
	}
	hostID, err := hostID(httpClient)
	if err != nil {
		return nil, err
	}
	sessionID, err := sessionID(httpClient, deviceID, hostID)
	if err != nil {
		return nil, err
	}

	body := authRequest{SessionID: sessionID, DeviceID: deviceID, HostID: hostID, RequestURL: requestURL}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling auth request: %w", err)
	}

	req, err := http.NewRequest("POST", authURL, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://data.usaswimming.org")
	req.Header.Set("Referer", "https://data.usaswimming.org/")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result authResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding auth response: %w", err)
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("empty access token in auth response")
	}
	return &authState{Token: result.AccessToken, DeviceID: deviceID, HostID: hostID}, nil
}

func sessionID(httpClient *http.Client, deviceID string, hostID string) (int64, error) {
	body := securityInfoRequest{
		AccessToken: "",
		Toxonomies:  []string{""},
		Scope:       "",
		UIProject:   "times-microsite-ui",
		BustCache:   false,
		AppName:     "Data",
		DeviceID:    deviceID,
		HostID:      hostID,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("marshaling security info request: %w", err)
	}

	req, err := http.NewRequest("POST", securityInfoURL, bytes.NewReader(b))
	if err != nil {
		return 0, fmt.Errorf("creating security info request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://data.usaswimming.org")
	req.Header.Set("Referer", "https://data.usaswimming.org/")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("security info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("security info request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result securityInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding security info response: %w", err)
	}
	if result.RequestID == "" {
		return 0, fmt.Errorf("empty requestId in security info response")
	}

	requestID, err := strconv.ParseInt(result.RequestID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing requestId %q: %w", result.RequestID, err)
	}
	return requestID * 13, nil
}

func getDeviceID(state *authState) (string, error) {
	if state != nil && state.DeviceID != "" {
		return state.DeviceID, nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("getting hostname: %w", err)
	}
	seed := fmt.Sprintf("swims|%s", hostname)
	var sum uint32
	for i := 0; i < len(seed); i++ {
		sum = sum*33 + uint32(seed[i])
	}
	return strconv.FormatUint(uint64(sum), 10), nil
}

func hostID(httpClient *http.Client) (string, error) {
	ip, err := publicIPv4(httpClient)
	if err != nil {
		return "", err
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return "", fmt.Errorf("parsing public IPv4 %q: %w", ip, err)
	}
	if !addr.Is4() {
		return "", fmt.Errorf("public IP %q is not IPv4", ip)
	}
	b := addr.As4()
	value := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(uint64(value), 10))), nil
}

func publicIPv4(httpClient *http.Client) (string, error) {
	for _, url := range publicIPURLs {
		resp, err := httpClient.Get(url)
		if err != nil {
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || readErr != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		addr, err := netip.ParseAddr(ip)
		if err == nil && addr.Is4() {
			return ip, nil
		}
	}

	return "", fmt.Errorf("getting public IPv4 address")
}
