package api

import (
	"context"
	"fmt"
	"net/url"
)

// PRListOptions configures the pull request listing.
type PRListOptions struct {
	RepoID    string // Repository ID (required for repo-scoped listing)
	Status    string // active, completed, abandoned, all (default: active)
	CreatorID string // Filter by creator (user ID)
	Top       int    // Max results (default: 20)
}

// ListPullRequests lists pull requests for a repository.
func (c *Client) ListPullRequests(ctx context.Context, opts PRListOptions) ([]PullRequest, error) {
	var path string
	if opts.RepoID != "" {
		path = fmt.Sprintf("_apis/git/repositories/%s/pullrequests", opts.RepoID)
	} else {
		path = "_apis/git/pullrequests"
	}

	u := c.projectURL(path)
	parsed, _ := url.Parse(u)
	q := parsed.Query()

	status := opts.Status
	if status == "" {
		status = "active"
	}
	q.Set("searchCriteria.status", status)

	if opts.CreatorID != "" {
		q.Set("searchCriteria.creatorId", opts.CreatorID)
	}

	top := opts.Top
	if top <= 0 {
		top = 20
	}
	q.Set("$top", fmt.Sprintf("%d", top))

	parsed.RawQuery = q.Encode()

	var result struct {
		Value []PullRequest `json:"value"`
	}
	if err := c.get(ctx, parsed.String(), &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetPullRequest fetches a single pull request.
func (c *Client) GetPullRequest(ctx context.Context, repoID string, prID int) (*PullRequest, error) {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests/%d", repoID, prID))
	var pr PullRequest
	if err := c.get(ctx, url, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePullRequest creates a new pull request.
func (c *Client) CreatePullRequest(ctx context.Context, repoID string, req CreatePRRequest) (*PullRequest, error) {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests", repoID))
	var pr PullRequest
	if err := c.post(ctx, url, req, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// ApprovePullRequest sets the current user's vote to Approved (10).
func (c *Client) ApprovePullRequest(ctx context.Context, repoID string, prID int, reviewerID string) error {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests/%d/reviewers/%s",
		repoID, prID, reviewerID))
	body := map[string]int{"vote": 10}
	return c.put(ctx, url, body, nil)
}

// CompletePullRequest completes (merges) a pull request.
func (c *Client) CompletePullRequest(ctx context.Context, repoID string, prID int) (*PullRequest, error) {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests/%d", repoID, prID))
	body := map[string]interface{}{
		"status": "completed",
		"completionOptions": map[string]interface{}{
			"deleteSourceBranch": true,
			"mergeStrategy":     "squash",
		},
	}
	resp, err := c.do(ctx, "PATCH", url, body, "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var pr PullRequest
	if err := decodeJSON(resp.Body, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// AbandonPullRequest sets the PR status to abandoned.
func (c *Client) AbandonPullRequest(ctx context.Context, repoID string, prID int) error {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests/%d", repoID, prID))
	body := map[string]string{"status": "abandoned"}
	_, err := c.do(ctx, "PATCH", url, body, "application/json")
	return err
}

// GetPullRequestWorkItems returns work items linked to a pull request.
func (c *Client) GetPullRequestWorkItems(ctx context.Context, repoID string, prID int) ([]WorkItem, error) {
	url := c.projectURL(fmt.Sprintf("_apis/git/repositories/%s/pullrequests/%d/workitems", repoID, prID))
	var result struct {
		Value []WIQLWorkItemRef `json:"value"`
	}
	if err := c.get(ctx, url, &result); err != nil {
		return nil, err
	}
	if len(result.Value) == 0 {
		return nil, nil
	}
	ids := make([]int, len(result.Value))
	for i, ref := range result.Value {
		ids[i] = ref.ID
	}
	return c.GetWorkItems(ctx, ids)
}
