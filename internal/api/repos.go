package api

import (
	"context"
	"encoding/json"
	"io"
)

// ListRepositories lists all git repositories in the project.
func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	u := c.projectURL("_apis/git/repositories")
	var result struct {
		Value []Repository `json:"value"`
	}
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetRepository fetches a single repository by name or ID.
func (c *Client) GetRepository(ctx context.Context, nameOrID string) (*Repository, error) {
	u := c.projectURL("_apis/git/repositories/" + nameOrID)
	var repo Repository
	if err := c.get(ctx, u, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

// decodeJSON is a helper to decode a response body into a destination.
func decodeJSON(r io.Reader, dest interface{}) error {
	return json.NewDecoder(r).Decode(dest)
}
