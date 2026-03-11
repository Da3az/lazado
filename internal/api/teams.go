package api

import (
	"context"
	"fmt"
	"net/url"
)

// TeamMember represents a member of a team.
type TeamMember struct {
	Identity IdentityRef `json:"identity"`
}

// GetTeamMembers fetches members of the default team in the project.
func (c *Client) GetTeamMembers(ctx context.Context) ([]string, error) {
	// First, find the default team for the project
	teamsURL := fmt.Sprintf("%s/_apis/projects/%s/teams?$top=1", c.baseURL, url.PathEscape(c.project))
	var teams struct {
		Value []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"value"`
	}
	if err := c.get(ctx, teamsURL, &teams); err != nil {
		return nil, err
	}
	if len(teams.Value) == 0 {
		return nil, fmt.Errorf("no teams found")
	}

	// Fetch members of the first team
	membersURL := fmt.Sprintf("%s/_apis/projects/%s/teams/%s/members",
		c.baseURL, url.PathEscape(c.project), url.PathEscape(teams.Value[0].ID))
	var result struct {
		Value []TeamMember `json:"value"`
	}
	if err := c.get(ctx, membersURL, &result); err != nil {
		return nil, err
	}

	names := make([]string, len(result.Value))
	for i, m := range result.Value {
		names[i] = m.Identity.DisplayName
	}
	return names, nil
}
