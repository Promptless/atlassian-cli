package present

import (
	"strings"
	"testing"

	"github.com/open-cli-collective/atlassian-go/present"
	"github.com/open-cli-collective/atlassian-go/testutil"

	"github.com/open-cli-collective/jira-ticket-cli/api"
)

func TestVersionDetail(t *testing.T) {
	t.Parallel()
	version := &api.Version{ID: "41699", Name: "Release 1", ProjectID: 10001, Released: true, ReleaseDate: "2026-10-22"}
	counts := &api.VersionIssueCounts{IssuesFixedCount: 3, IssuesAffectedCount: 1, IssueCountWithCustomFieldsShowingVersion: 0}
	model := VersionPresenter{}.PresentDetail(version, counts, false, false)
	want := &present.OutputModel{Sections: []present.Section{
		msg("41699  Release 1"),
		msg("Project ID: 10001   Released: yes   Archived: no"),
		msg("Start date: -   Release date: 2026-10-22"),
		msg("Fix version issues: 3   Affected version issues: 1   Custom field issues: 0"),
	}}
	testutil.Equal(t, model, want)
}

func TestVersionDescriptionModes(t *testing.T) {
	t.Parallel()
	version := &api.Version{ID: "41699", Name: "Release 1", ProjectID: 10001, Description: strings.Repeat("a", 600), Self: "https://example.atlassian.net/rest/api/3/version/41699"}
	counts := &api.VersionIssueCounts{}
	for _, tt := range []struct {
		name               string
		extended, fulltext bool
		wantDescription    string
	}{
		{"default truncation", false, false, strings.Repeat("a", 500) + "\n[truncated — use --fulltext for complete body]"},
		{"full text", false, true, version.Description},
		{"extended includes full text", true, false, version.Description},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := VersionPresenter{}.PresentDetail(version, counts, tt.extended, tt.fulltext)
			last := model.Sections[len(model.Sections)-1].(*present.MessageSection)
			testutil.Equal(t, last.Message, tt.wantDescription)
			if tt.extended {
				testutil.Equal(t, model.Sections[4], present.Section(msg("URL: "+version.Self)))
			}
		})
	}
}

func TestVersionDetailProjection(t *testing.T) {
	t.Parallel()
	version := &api.Version{ID: "41699", Name: "Release 1", ProjectID: 10001}
	model := VersionPresenter{}.PresentDetailProjection(version, &api.VersionIssueCounts{}, false)
	detail := model.Sections[0].(*present.DetailSection)
	testutil.Equal(t, detail.Fields, []present.Field{
		{Label: "ID", Value: "41699"},
		{Label: "NAME", Value: "Release 1"},
		{Label: "PROJECT_ID", Value: "10001"},
		{Label: "RELEASED", Value: "no"},
		{Label: "ARCHIVED", Value: "no"},
		{Label: "START_DATE", Value: "-"},
		{Label: "RELEASE_DATE", Value: "-"},
		{Label: "FIXED_ISSUES", Value: "0"},
		{Label: "AFFECTED_ISSUES", Value: "0"},
		{Label: "CUSTOM_FIELD_ISSUES", Value: "0"},
		{Label: "DESCRIPTION", Value: "-"},
		{Label: "URL", Value: "-"},
	})
}
