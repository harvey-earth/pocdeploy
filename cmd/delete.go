package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/harvey-earth/pocdeploy/internal"
)

// destroyCmd represents the destroy command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes kubernetes cluster",
	Long: `deletes the kubernetes cluster created with the "create" command.

This will delete resources using Terraform, and it will not be graceful.`,
	Example: `pocdeploy delete -t [kind]`,
	Run: func(cmd *cobra.Command, args []string) {

		// Run DeleteKindCluster for type kind
		if viper.GetString("type") == "kind" {
			if err := internal.DeleteKindCluster(viper.GetString("name")); err != nil {
				internal.Error("Error deleting Kind cluster named 'test'", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
