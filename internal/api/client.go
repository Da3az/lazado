package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/da3az/lazado/internal/auth"
)

// Client communicates with the Azure DevOps REST API.
type Client struct {
	httpClient *http.Client
	baseURL    string // e.g. "https://dev.azure.com/MyOrg"
	project    string
	auth       auth.Provider
	userID     string // cached authenticated user ID
}

// NewClient creates an API client for the given organization and project.
func NewClient(baseURL, project string, authProvider auth.Provider) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		project:    project,
		auth:       authProvider,
	}
}

// SetUserID sets the cached user ID (obtained during init/connection validation).
func (c *Client) SetUserID(id string) {
	c.userID = id
}

// UserID returns the cached authenticated user ID.
func (c *Client) UserID() string {
	return c.userID
}

// BaseURL returns the base organization URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Project returns the project name.
func (c *Client) Project() string {
	return c.project
}

// ValidateConnection checks if the PAT is valid by calling _apis/connectionData.
func (c *Client) ValidateConnection(ctx context.Context) (*ConnectionData, error) {
	var result ConnectionData
	err := c.get(ctx, c.baseURL+"/_apis/connectionData", &result)
	if err != nil {
		return nil, fmt.Errorf("connection validation failed: %w", err)
	}
	return &result, nil
}

const apiVersion = "7.1-preview"

func (c *Client) projectURL(path string) string {
	return fmt.Sprintf("%s/%s/%s", c.baseURL, c.project, path)
}

func (c *Client) addAuth(req *http.Request) error {
	token, err := c.auth.Token()
	if err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + token))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

func (c *Client) do(ctx context.Context, method, url string, body interface{}, contentType string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if err := c.addAuth(req); err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add API version as query parameter if not already present
	q := req.URL.Query()
	if q.Get("api-version") == "" {
		q.Set("api-version", apiVersion)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return resp, nil
}

func (c *Client) get(ctx context.Context, url string, dest interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, url, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *Client) post(ctx context.Context, url string, body, dest interface{}) error {
	resp, err := c.do(ctx, http.MethodPost, url, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

func (c *Client) patch(ctx context.Context, url string, body, dest interface{}) error {
	resp, err := c.do(ctx, http.MethodPatch, url, body, "application/json-patch+json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

func (c *Client) put(ctx context.Context, url string, body, dest interface{}) error {
	resp, err := c.do(ctx, http.MethodPut, url, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if dest != nil {
		return json.NewDecoder(resp.Body).Decode(dest)
	}
	return nil
}

// APIError represents an HTTP error from the Azure DevOps API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Azure DevOps API error (HTTP %d): %s", e.StatusCode, e.Message)
}
