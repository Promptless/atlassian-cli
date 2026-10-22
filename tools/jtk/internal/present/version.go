package present

import (
	"fmt"

	"github.com/open-cli-collective/atlassian-go/present"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/open-cli-collective/jira-ticket-cli/internal/present/projection"
)

// VersionPresenter creates presentation models for version metadata and counts.
type VersionPresenter struct{}

// VersionDetailSpec declares the fields available on versions get.
var VersionDetailSpec = projection.Registry{
	{Header: "ID", Identity: true},
	{Header: "NAME"},
	{Header: "PROJECT_ID"},
	{Header: "RELEASED"},
	{Header: "ARCHIVED"},
	{Header: "START_DATE"},
	{Header: "RELEASE_DATE"},
	{Header: "FIXED_ISSUES"},
	{Header: "AFFECTED_ISSUES"},
	{Header: "CUSTOM_FIELD_ISSUES"},
	{Header: "DESCRIPTION"},
	{Header: "URL", Extended: true},
}

// PresentDetail shows release metadata and issue counts, including explicit zeros.
func (VersionPresenter) PresentDetail(version *api.Version, counts *api.VersionIssueCounts, extended, fulltext bool) *present.OutputModel {
	sections := []present.Section{
		msg(fmt.Sprintf("%s  %s", version.ID, version.Name)),
		msg(fmt.Sprintf("Project ID: %d   Released: %s   Archived: %s", version.ProjectID, BoolString(version.Released), BoolString(version.Archived))),
		msg(fmt.Sprintf("Start date: %s   Release date: %s", OrDash(version.StartDate), OrDash(version.ReleaseDate))),
		msg(fmt.Sprintf("Fix version issues: %d   Affected version issues: %d   Custom field issues: %d", counts.IssuesFixedCount, counts.IssuesAffectedCount, counts.IssueCountWithCustomFieldsShowingVersion)),
	}
	if extended {
		sections = append(sections, msg("URL: "+OrDash(version.Self)))
	}
	if version.Description != "" {
		sections = append(sections, msg(""), msg("Description:"), msg(versionDescription(version.Description, extended || fulltext)))
	}
	return &present.OutputModel{Sections: sections}
}

// PresentDetailProjection builds the labeled values selected with --fields.
func (VersionPresenter) PresentDetailProjection(version *api.Version, counts *api.VersionIssueCounts, fulltext bool) *present.OutputModel {
	fields := []present.Field{
		{Label: "ID", Value: version.ID},
		{Label: "NAME", Value: version.Name},
		{Label: "PROJECT_ID", Value: FormatInt(version.ProjectID)},
		{Label: "RELEASED", Value: BoolString(version.Released)},
		{Label: "ARCHIVED", Value: BoolString(version.Archived)},
		{Label: "START_DATE", Value: OrDash(version.StartDate)},
		{Label: "RELEASE_DATE", Value: OrDash(version.ReleaseDate)},
		{Label: "FIXED_ISSUES", Value: fmt.Sprintf("%d", counts.IssuesFixedCount)},
		{Label: "AFFECTED_ISSUES", Value: fmt.Sprintf("%d", counts.IssuesAffectedCount)},
		{Label: "CUSTOM_FIELD_ISSUES", Value: fmt.Sprintf("%d", counts.IssueCountWithCustomFieldsShowingVersion)},
		{Label: "DESCRIPTION", Value: OrDash(versionDescription(version.Description, fulltext))},
		{Label: "URL", Value: OrDash(version.Self)},
	}
	return &present.OutputModel{Sections: []present.Section{&present.DetailSection{Fields: fields}}}
}

func versionDescription(description string, fulltext bool) string {
	const maxDescriptionLength = 500
	runes := []rune(description)
	if fulltext || len(runes) <= maxDescriptionLength {
		return description
	}
	return string(runes[:maxDescriptionLength]) + "\n[truncated — use --fulltext for complete body]"
}
