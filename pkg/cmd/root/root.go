package root

import (
	"github.com/spf13/cobra"

	"github.com/atlanticbt/magecli/pkg/cmd/api"
	"github.com/atlanticbt/magecli/pkg/cmd/attribute"
	"github.com/atlanticbt/magecli/pkg/cmd/auth"
	"github.com/atlanticbt/magecli/pkg/cmd/category"
	"github.com/atlanticbt/magecli/pkg/cmd/cms"
	configcmd "github.com/atlanticbt/magecli/pkg/cmd/config"
	contextcmd "github.com/atlanticbt/magecli/pkg/cmd/context"
	"github.com/atlanticbt/magecli/pkg/cmd/inventory"
	"github.com/atlanticbt/magecli/pkg/cmd/product"
	"github.com/atlanticbt/magecli/pkg/cmd/promo"
	"github.com/atlanticbt/magecli/pkg/cmd/store"
	"github.com/atlanticbt/magecli/pkg/cmd/update"
	"github.com/atlanticbt/magecli/pkg/cmdutil"
)

func NewCmdRoot(f *cmdutil.Factory) (*cobra.Command, error) {
	ios, err := f.Streams()
	if err != nil {
		return nil, err
	}

	root := &cobra.Command{
		Use:   f.ExecutableName,
		Short: "Magento 2 CLI for AI agents and developers.",
		Long: `Query Magento 2 stores via the REST API from the command line.

Common flows:
  magecli auth login https://store.example.com --token <token>
  magecli product list --filter "name like %shirt%" --json
  magecli category tree --json`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	root.PersistentFlags().StringP("context", "c", "", "Active Magento context name")
	root.PersistentFlags().String("store-code", "", "Override Magento store code")
	root.PersistentFlags().Bool("json", false, "Output in JSON format")
	root.PersistentFlags().Bool("yaml", false, "Output in YAML format")
	root.PersistentFlags().String("jq", "", "Apply a jq expression to JSON output (requires --json)")
	root.PersistentFlags().String("template", "", "Render output using Go templates")

	root.AddCommand(
		auth.NewCmdAuth(f),
		contextcmd.NewCmdContext(f),
		product.NewCmdProduct(f),
		category.NewCmdCategory(f),
		attribute.NewCmdAttribute(f),
		inventory.NewCmdInventory(f),
		store.NewCmdStore(f),
		configcmd.NewCmdConfig(f),
		promo.NewCmdPromo(f),
		cms.NewCmdCMS(f),
		api.NewCmdAPI(f),
		update.NewCmdUpdate(f),
	)

	root.Version = f.AppVersion
	root.SetIn(ios.In)
	root.SetOut(ios.Out)
	root.SetErr(ios.ErrOut)

	return root, nil
}
