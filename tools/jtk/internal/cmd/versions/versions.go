package versions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/open-cli-collective/jira-ticket-cli/api"
	"github.com/open-cli-collective/jira-ticket-cli/internal/cmd/root"
	jtkpresent "github.com/open-cli-collective/jira-ticket-cli/internal/present"
	"github.com/open-cli-collective/jira-ticket-cli/internal/present/projection"
)

// Register adds read-only Jira version research commands.
func Register(parent *cobra.Command, opts *root.Options) {
	cmd := &cobra.Command{
		Use: "versions", Aliases: []string{"ver"},
		Short: "Read Jira project versions",
	}
	cmd.AddCommand(newGetCmd(opts))
	parent.AddCommand(cmd)
}

func newGetCmd(opts *root.Options) *cobra.Command {
	var fieldsFlag string
	cmd := &cobra.Command{
		Use: "get <version-id>", Short: "Get release metadata and related issue counts",
		Long:    "Read a Jira version by its numeric ID, including its project, release state, and fixed/affected/custom-field issue counts. A not-found response can mean the version is absent or inaccessible to the connected account.",
		Example: "  jtk versions get 10000\n  jtk versions get 10000 --fields NAME,PROJECT_ID,FIXED_ISSUES\n  jtk versions get 10000 --extended\n  jtk versions get 10000 --id",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), opts, args[0], fieldsFlag)
		},
	}
	cmd.Flags().StringVar(&fieldsFlag, "fields", "", "Comma-separated display fields (VersionDetailSpec headers)")
	return cmd
}

func runGet(ctx context.Context, opts *root.Options, id, fieldsFlag string) error {
	client, err := opts.APIClient()
	if err != nil {
		return err
	}
	if opts.EmitIDOnly() {
		version, err := client.GetVersion(ctx, id)
		if err != nil {
			return err
		}
		return jtkpresent.EmitIDs(opts, []string{version.ID})
	}
	selected, projected, err := projection.Resolve(ctx, jtkpresent.VersionDetailSpec, opts.IsExtended(), fieldsFlag, noFieldFetch, "versions get")
	if err != nil {
		return err
	}
	version, err := client.GetVersion(ctx, id)
	if err != nil {
		return err
	}
	counts, err := client.GetVersionIssueCounts(ctx, id)
	if err != nil {
		return err
	}
	presenter := jtkpresent.VersionPresenter{}
	if projected {
		model := presenter.PresentDetailProjection(version, counts, opts.IsFullText() || opts.IsExtended())
		projection.ApplyToDetailInModel(model, selected)
		return jtkpresent.Emit(opts, model)
	}
	return jtkpresent.Emit(opts, presenter.PresentDetail(version, counts, opts.IsExtended(), opts.IsFullText()))
}

func noFieldFetch(context.Context) ([]api.Field, error) {
	return nil, nil
}
