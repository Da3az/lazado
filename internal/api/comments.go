package api

import (
	"context"
	"fmt"
)

// GetWorkItemComments fetches comments for a work item.
func (c *Client) GetWorkItemComments(ctx context.Context, wiID int) ([]Comment, error) {
	u := c.projectURL(fmt.Sprintf("_apis/wit/workitems/%d/comments", wiID))
	var result struct {
		Comments []Comment `json:"comments"`
	}
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	return result.Comments, nil
}

// AddWorkItemComment posts a new comment on a work item.
func (c *Client) AddWorkItemComment(ctx context.Context, wiID int, text string) (*Comment, error) {
	u := c.projectURL(fmt.Sprintf("_apis/wit/workitems/%d/comments", wiID))
	body := map[string]string{"text": text}
	var comment Comment
	if err := c.post(ctx, u, body, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}
