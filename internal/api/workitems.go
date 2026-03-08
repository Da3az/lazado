package api

import (
	"context"
	"fmt"
	"strings"
)

// GetWorkItem fetches a single work item by ID.
func (c *Client) GetWorkItem(ctx context.Context, id int) (*WorkItem, error) {
	url := c.projectURL(fmt.Sprintf("_apis/wit/workitems/%d?$expand=all", id))
	var wi WorkItem
	if err := c.get(ctx, url, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// GetWorkItems fetches multiple work items by their IDs.
func (c *Client) GetWorkItems(ctx context.Context, ids []int) ([]WorkItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// API supports up to 200 IDs per request
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	url := c.projectURL(fmt.Sprintf("_apis/wit/workitems?ids=%s&$expand=all", strings.Join(idStrs, ",")))

	var result struct {
		Value []WorkItem `json:"value"`
	}
	if err := c.get(ctx, url, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// QueryWorkItems executes a WIQL query and returns the matching work items.
func (c *Client) QueryWorkItems(ctx context.Context, wiql string) ([]WorkItem, error) {
	url := c.projectURL("_apis/wit/wiql")
	body := map[string]string{"query": wiql}

	var result WIQLResult
	if err := c.post(ctx, url, body, &result); err != nil {
		return nil, err
	}

	if len(result.WorkItems) == 0 {
		return nil, nil
	}

	// Fetch full work item details
	ids := make([]int, len(result.WorkItems))
	for i, ref := range result.WorkItems {
		ids[i] = ref.ID
	}
	return c.GetWorkItems(ctx, ids)
}

// CreateWorkItem creates a new work item of the specified type.
func (c *Client) CreateWorkItem(ctx context.Context, wiType string, ops []PatchOperation) (*WorkItem, error) {
	url := c.projectURL(fmt.Sprintf("_apis/wit/workitems/$%s", wiType))
	var wi WorkItem
	if err := c.patch(ctx, url, ops, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// UpdateWorkItem updates a work item with the given patch operations.
func (c *Client) UpdateWorkItem(ctx context.Context, id int, ops []PatchOperation) (*WorkItem, error) {
	url := c.projectURL(fmt.Sprintf("_apis/wit/workitems/%d", id))
	var wi WorkItem
	if err := c.patch(ctx, url, ops, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}
