package versions

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-cli-collective/atlassian-go/testutil"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/open-cli-collective/jira-ticket-cli/internal/cmd/root"
)

func TestRegister(t *testing.T) {
	t.Parallel()
	rootCmd, opts := root.NewCmd()
	Register(rootCmd, opts)
	cmd, _, err := rootCmd.Find([]string{"ver", "get"})
	testutil.RequireNoError(t, err)
	testutil.Equal(t, cmd.Name(), "get")
	testutil.NotNil(t, cmd.Flags().Lookup("fields"))
	testutil.Error(t, cmd.ValidateArgs(nil))
	testutil.Error(t, cmd.ValidateArgs([]string{"41699", "41702"}))
}

func TestRunGet(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, fields, wantOutput string
		idOnly, extended         bool
		countStatus              int
		wantErr                  bool
		wantPaths                []string
	}{
		{
			name: "metadata and counts",
			wantOutput: "41699  Release 1\nProject ID: 10001   Released: yes   Archived: no\n" +
				"Start date: -   Release date: 2026-10-22\nFix version issues: 0   Affected version issues: 0   Custom field issues: 0\n",
			countStatus: http.StatusOK,
			wantPaths:   []string{"/rest/api/3/version/41699", "/rest/api/3/version/41699/relatedIssueCounts"},
		},
		{
			name: "identifier ignores projection and counts", fields: "NOT_A_FIELD", idOnly: true, extended: true,
			wantOutput: "41699\n", countStatus: http.StatusForbidden,
			wantPaths: []string{"/rest/api/3/version/41699"},
		},
		{
			name: "field projection", fields: "NAME,PROJECT_ID,FIXED_ISSUES",
			wantOutput: "ID: 41699\nNAME: Release 1\nPROJECT_ID: 10001\nFIXED_ISSUES: 0\n", countStatus: http.StatusOK,
			wantPaths: []string{"/rest/api/3/version/41699", "/rest/api/3/version/41699/relatedIssueCounts"},
		},
		{
			name: "invalid fields make no requests", fields: "NOT_A_FIELD", countStatus: http.StatusOK, wantErr: true,
		},
		{
			name: "counts error does not print zero", countStatus: http.StatusForbidden, wantErr: true,
			wantPaths: []string{"/rest/api/3/version/41699", "/rest/api/3/version/41699/relatedIssueCounts"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var paths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				switch r.URL.Path {
				case "/rest/api/3/version/41699":
					_, _ = w.Write([]byte(`{"id":"41699","name":"Release 1","projectId":10001,"released":true,"releaseDate":"2026-10-22"}`))
				case "/rest/api/3/version/41699/relatedIssueCounts":
					w.WriteHeader(tt.countStatus)
					if tt.countStatus == http.StatusOK {
						_, _ = w.Write([]byte(`{"issuesFixedCount":0,"issuesAffectedCount":0,"issueCountWithCustomFieldsShowingVersion":0}`))
					}
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client, err := api.New(api.ClientConfig{URL: server.URL, Email: "test@example.com", APIToken: "test-token"})
			testutil.RequireNoError(t, err)
			var stdout, stderr bytes.Buffer
			opts := &root.Options{NoColor: true, IDOnly: tt.idOnly, Extended: tt.extended, Stdout: &stdout, Stderr: &stderr}
			opts.SetAPIClient(client)
			err = runGet(context.Background(), opts, "41699", tt.fields)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.RequireNoError(t, err)
			}
			testutil.Equal(t, stdout.String(), tt.wantOutput)
			testutil.Equal(t, stderr.String(), "")
			testutil.Equal(t, paths, tt.wantPaths)
		})
	}
}
