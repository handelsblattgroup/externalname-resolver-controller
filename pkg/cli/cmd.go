package cli

import (
	"fmt"

	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/cli/version"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/cli/watch"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "externalname-resolver-controller",
	Short: "",
	Long:  `Kubernetes controller to manage ExternalName type Services resolving their DNS entry, generating and keeping in sync specific Endpoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("please use a subcommand...")
		cmd.Usage()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(version.Command())
	rootCmd.AddCommand(watch.Command())
}

func Command() *cobra.Command {
	return rootCmd
}
