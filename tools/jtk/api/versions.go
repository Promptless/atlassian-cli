package api //nolint:revive // package name is intentional

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Version represents a Jira project version and its release metadata.
type Version struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ProjectID   int    `json:"projectId"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	StartDate   string `json:"startDate,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Description string `json:"description,omitempty"`
	Self        string `json:"self,omitempty"`
}

// VersionIssueCounts contains Jira's counts for each way issues use a version.
type VersionIssueCounts struct {
	IssuesFixedCount                         int `json:"issuesFixedCount"`
	IssuesAffectedCount                      int `json:"issuesAffectedCount"`
	IssueCountWithCustomFieldsShowingVersion int `json:"issueCountWithCustomFieldsShowingVersion"`
}

// GetVersion reads a version by ID. A 404 can mean absent or inaccessible.
// Both basic and OAuth bearer authentication use this Jira Platform endpoint.
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-project-versions/#api-rest-api-3-version-id-get
func (c *Client) GetVersion(ctx context.Context, id string) (*Version, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("version ID is required")
	}
	body, err := c.Get(ctx, fmt.Sprintf("%s/version/%s", c.BaseURL, url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("getting version %s: %w", id, err)
	}
	var version Version
	if err := json.Unmarshal(body, &version); err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}
	if version.ID != id || strings.TrimSpace(version.Name) == "" || version.ProjectID <= 0 {
		return nil, fmt.Errorf("incomplete or mismatched Jira metadata for version %s", id)
	}
	return &version, nil
}

// GetVersionIssueCounts reads fixed, affected, and custom-field issue counts.
// Missing counts are errors rather than evidence that the version is empty.
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-project-versions/#api-rest-api-3-version-id-relatedissuecounts-get
func (c *Client) GetVersionIssueCounts(ctx context.Context, id string) (*VersionIssueCounts, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("version ID is required")
	}
	body, err := c.Get(ctx, fmt.Sprintf("%s/version/%s/relatedIssueCounts", c.BaseURL, url.PathEscape(id)))
	if err != nil {
		return nil, fmt.Errorf("getting issue counts for version %s: %w", id, err)
	}
	var response struct {
		Fixed    *int `json:"issuesFixedCount"`
		Affected *int `json:"issuesAffectedCount"`
		Custom   *int `json:"issueCountWithCustomFieldsShowingVersion"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("parsing version issue counts: %w", err)
	}
	if response.Fixed == nil || response.Affected == nil || response.Custom == nil {
		return nil, fmt.Errorf("incomplete Jira issue counts for version %s", id)
	}
	if *response.Fixed < 0 || *response.Affected < 0 || *response.Custom < 0 {
		return nil, fmt.Errorf("negative Jira issue counts for version %s", id)
	}
	return &VersionIssueCounts{
		IssuesFixedCount:                         *response.Fixed,
		IssuesAffectedCount:                      *response.Affected,
		IssueCountWithCustomFieldsShowingVersion: *response.Custom,
	}, nil
}
