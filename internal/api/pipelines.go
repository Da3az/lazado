package api

import (
	"context"
	"fmt"
	"net/url"
)

// ListPipelineRuns lists recent pipeline runs.
func (c *Client) ListPipelineRuns(ctx context.Context, top int) ([]PipelineRun, error) {
	if top <= 0 {
		top = 20
	}
	u := c.projectURL("_apis/pipelines/runs")
	parsed, _ := url.Parse(u)
	q := parsed.Query()
	q.Set("$top", fmt.Sprintf("%d", top))
	parsed.RawQuery = q.Encode()

	var result struct {
		Value []PipelineRun `json:"value"`
	}
	if err := c.get(ctx, parsed.String(), &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// ListPipelines lists pipeline definitions.
func (c *Client) ListPipelines(ctx context.Context) ([]Pipeline, error) {
	u := c.projectURL("_apis/pipelines")
	var result struct {
		Value []Pipeline `json:"value"`
	}
	if err := c.get(ctx, u, &result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

// RunPipeline triggers a new run of the specified pipeline.
func (c *Client) RunPipeline(ctx context.Context, pipelineID int) (*PipelineRun, error) {
	u := c.projectURL(fmt.Sprintf("_apis/pipelines/%d/runs", pipelineID))
	body := map[string]interface{}{}
	var run PipelineRun
	if err := c.post(ctx, u, body, &run); err != nil {
		return nil, err
	}
	return &run, nil
}
