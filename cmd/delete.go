/*
Copyright © 2024 Jamie Harvey <jamie@harvey.earth>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	c "github.com/harvey-earth/pocdeploy/internal"
)

// destroyCmd represents the destroy command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes kubernetes cluster",
	Long: `deletes the kubernetes cluster created with the "create" command.

This will delete resources using Terraform, and it will not be graceful.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("delete called")

		if viper.GetString("type") == "kind" {
			c.DeleteKindCluster("test")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
