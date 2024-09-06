package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	c "github.com/harvey-earth/pocdeploy/internal"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a kubernetes cluster and deploy",
	Long: `creates a Kind Kubernetes cluster and deploys a frontend application with a CloudNative PG backend.

If the type is not set, the default is a local Kind cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")

		// Create cluster
		if viper.GetString("type") == "kind" {
			c.CreateKindCluster()
		}

		// Build image using name and version from config and return them
		imgName, imgVers := c.BuildImage()

		// Install CloudNativePG Operator
		c.InstallBackend()

		// Load docker image
		if viper.GetString("type") == "kind" {
			c.LoadKindImage(imgName, imgVers)
		}

		// Create namespaces
		c.CreateNamespaces()

		// Deploy frontend with generated secret key
		c.CreateSecretKeySecret()
		c.ConfigureFrontend()

		// Configure CloudNativePG
		c.ConfigureBackend()

		// Run Django migrations
		c.InitBackend()

		// Deploy prometheus
		c.ConfigureMonitoring()

		// Create job that creates superuser
		c.CreateAdminUser()
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

}
