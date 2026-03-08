package api

import (
	"context"
	"fmt"
)

// GetWorkItemTypeStates fetches the available states for a work item type.
func (c *Client) GetWorkItemTypeStates(ctx context.Context, wiType string) ([]WorkItemTypeState, error) {
	u := c.projectURL(fmt.Sprintf("_apis/wit/workitemtypes/%s/states", wiType))
	var result struct {
		Value []WorkItemTypeState `json:"value"`
	}
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// GetWorkItemTypes lists available work item types in the project.
func (c *Client) GetWorkItemTypes(ctx context.Context) ([]string, error) {
	u := c.projectURL("_apis/wit/workitemtypes")
	var result struct {
		Value []struct {
			Name string `json:"name"`
		} `json:"value"`
	}
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	types := make([]string, len(result.Value))
	for i, t := range result.Value {
		types[i] = t.Name
	}
	return types, nil
}

// DefaultStates returns built-in fallback states for common work item types.
// These match the Agile process template defaults.
func DefaultStates() map[string][]WorkItemTypeState {
	return map[string][]WorkItemTypeState{
		"Task": {
			{Name: "New", Category: "Proposed"},
			{Name: "Active", Category: "InProgress"},
			{Name: "Closed", Category: "Completed"},
			{Name: "Removed", Category: "Removed"},
		},
		"User Story": {
			{Name: "New", Category: "Proposed"},
			{Name: "Active", Category: "InProgress"},
			{Name: "Resolved", Category: "Resolved"},
			{Name: "Closed", Category: "Completed"},
			{Name: "Removed", Category: "Removed"},
		},
		"Bug": {
			{Name: "New", Category: "Proposed"},
			{Name: "Active", Category: "InProgress"},
			{Name: "Resolved", Category: "Resolved"},
			{Name: "Closed", Category: "Completed"},
		},
		"Epic": {
			{Name: "New", Category: "Proposed"},
			{Name: "Active", Category: "InProgress"},
			{Name: "Resolved", Category: "Resolved"},
			{Name: "Closed", Category: "Completed"},
		},
		"Feature": {
			{Name: "New", Category: "Proposed"},
			{Name: "Active", Category: "InProgress"},
			{Name: "Resolved", Category: "Resolved"},
			{Name: "Closed", Category: "Completed"},
		},
	}
}
