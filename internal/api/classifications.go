package api

import (
	"context"
)

// ClassificationNode represents an area or iteration path node.
type ClassificationNode struct {
	ID       int                  `json:"id"`
	Name     string               `json:"name"`
	Path     string               `json:"path"`
	Children []ClassificationNode `json:"children"`
}

// GetAreaPaths fetches the area path tree and returns flattened paths.
func (c *Client) GetAreaPaths(ctx context.Context) ([]string, error) {
	u := c.projectURL("_apis/wit/classificationnodes/areas?$depth=10")
	var root ClassificationNode
	if err := c.get(ctx, u, &root); err != nil {
		return nil, err
	}
	var paths []string
	flattenPaths(&paths, root, c.project)
	return paths, nil
}

// GetIterationPaths fetches the iteration path tree and returns flattened paths.
func (c *Client) GetIterationPaths(ctx context.Context) ([]string, error) {
	u := c.projectURL("_apis/wit/classificationnodes/iterations?$depth=10")
	var root ClassificationNode
	if err := c.get(ctx, u, &root); err != nil {
		return nil, err
	}
	var paths []string
	flattenPaths(&paths, root, c.project)
	return paths, nil
}

func flattenPaths(paths *[]string, node ClassificationNode, prefix string) {
	*paths = append(*paths, prefix)
	for _, child := range node.Children {
		flattenPaths(paths, child, prefix+"\\"+child.Name)
	}
}
