package xui

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

// apiResponse is the standard x-ui API response format.
type apiResponse struct {
	Success bool            `json:"success"`
	Msg     string          `json:"msg"`
	Obj     json.RawMessage `json:"obj"`
}

// Client is an authenticated HTTP client for the x-ui panel API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new x-ui API client.
func NewClient(panelInfo *PanelInfo) *Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // localhost, self-signed certs
		},
	}
	return &Client{
		baseURL: strings.TrimSuffix(panelInfo.BaseURL, "/"),
		httpClient: &http.Client{
			Jar:       jar,
			Transport: transport,
		},
	}
}

// Login authenticates with the x-ui panel.
func (c *Client) Login(username, password string) error {
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	resp, err := c.postForm("login", data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("login failed: %s", resp.Msg)
	}
	return nil
}

// postForm sends a form-encoded POST request to the given path.
func (c *Client) postForm(path string, data url.Values) (*apiResponse, error) {
	reqURL := c.baseURL + "/" + path
	resp, err := c.httpClient.PostForm(reqURL, data)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("endpoint not found: %s (check panel base path)", path)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(body))
	}
	return &apiResp, nil
}

// post sends a POST request with the given body.
func (c *Client) post(path string) (*apiResponse, error) {
	return c.postForm(path, nil)
}

// get sends a GET request to the given path.
func (c *Client) get(path string) (*apiResponse, error) {
	reqURL := c.baseURL + "/" + path
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &apiResp, nil
}
