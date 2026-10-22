package api //nolint:revive // package name is intentional

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-cli-collective/atlassian-go/testutil"
)

func TestGetVersion(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body string
		status     int
		wantErr    bool
	}{
		{"version metadata", `{"id":"41699","name":"Release 1","projectId":10001,"released":true,"archived":false,"releaseDate":"2026-10-22","description":"A release","self":"https://example.atlassian.net/rest/api/3/version/41699"}`, http.StatusOK, false},
		{"not found or inaccessible", `{"errorMessages":["Version not found"]}`, http.StatusNotFound, true},
		{"unauthorized", `{"errorMessages":["Unauthorized"]}`, http.StatusUnauthorized, true},
		{"invalid json", `{`, http.StatusOK, true},
		{"empty success", `{}`, http.StatusOK, true},
		{"different version", `{"id":"41702","name":"Release 1","projectId":10001}`, http.StatusOK, true},
		{"missing name", `{"id":"41699","projectId":10001}`, http.StatusOK, true},
		{"missing project", `{"id":"41699","name":"Release 1"}`, http.StatusOK, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				testutil.Equal(t, r.Method, http.MethodGet)
				testutil.Equal(t, r.URL.Path, "/rest/api/3/version/41699")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			client, err := New(ClientConfig{URL: server.URL, Email: "test@example.com", APIToken: "test-token"})
			testutil.RequireNoError(t, err)
			version, err := client.GetVersion(context.Background(), "41699")
			if tt.wantErr {
				testutil.Error(t, err)
				return
			}
			testutil.RequireNoError(t, err)
			testutil.Equal(t, version.ID, "41699")
			testutil.Equal(t, version.Name, "Release 1")
			testutil.Equal(t, version.ProjectID, 10001)
			testutil.True(t, version.Released)
			testutil.Equal(t, version.ReleaseDate, "2026-10-22")
		})
	}
}

func TestGetVersionIssueCounts(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body string
		status     int
		wantErr    bool
		wantFixed  int
	}{
		{"populated version", `{"issuesFixedCount":3,"issuesAffectedCount":1,"issueCountWithCustomFieldsShowingVersion":2}`, http.StatusOK, false, 3},
		{"empty version", `{"issuesFixedCount":0,"issuesAffectedCount":0,"issueCountWithCustomFieldsShowingVersion":0}`, http.StatusOK, false, 0},
		{"not found or inaccessible", `{"errorMessages":["Version not found"]}`, http.StatusNotFound, true, 0},
		{"invalid json", `{`, http.StatusOK, true, 0},
		{"missing counts", `{}`, http.StatusOK, true, 0},
		{"partial counts", `{"issuesFixedCount":0,"issuesAffectedCount":0}`, http.StatusOK, true, 0},
		{"negative count", `{"issuesFixedCount":-1,"issuesAffectedCount":0,"issueCountWithCustomFieldsShowingVersion":0}`, http.StatusOK, true, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				testutil.Equal(t, r.Method, http.MethodGet)
				testutil.Equal(t, r.URL.Path, "/rest/api/3/version/41699/relatedIssueCounts")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			client, err := New(ClientConfig{URL: server.URL, Email: "test@example.com", APIToken: "test-token"})
			testutil.RequireNoError(t, err)
			counts, err := client.GetVersionIssueCounts(context.Background(), "41699")
			if tt.wantErr {
				testutil.Error(t, err)
				return
			}
			testutil.RequireNoError(t, err)
			testutil.Equal(t, counts.IssuesFixedCount, tt.wantFixed)
		})
	}
}

func TestVersionReadsRequireID(t *testing.T) {
	t.Parallel()
	client, err := New(ClientConfig{URL: "https://example.atlassian.net", Email: "test@example.com", APIToken: "test-token"})
	testutil.RequireNoError(t, err)
	_, err = client.GetVersion(context.Background(), "")
	testutil.Error(t, err)
	_, err = client.GetVersionIssueCounts(context.Background(), "")
	testutil.Error(t, err)
}
