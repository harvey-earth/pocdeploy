package cmd

import (
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
		imgName, imgVers, err := c.BuildImage()
		if err != nil {
			c.Error("Error building image", err)
		}

		// Install CloudNativePG Operator
		err = c.InstallBackend()
		if err != nil {
			c.Error("Error installing backend", err)
		}

		// Load docker image
		if viper.GetString("type") == "kind" {
			err = c.LoadKindImage(imgName, imgVers)
			if err != nil {
				c.Error("Error loading image to Kind", err)
			}
		}

		// Create namespaces
		err = c.CreateNamespaces()
		if err != nil {
			c.Error("Error creating namespaces", err)
		}

		// Deploy frontend with generated secret key
		err = c.CreateSecretKeySecret()
		if err != nil {
			c.Error("Error creating namespaces", err)
		}
		err = c.ConfigureFrontend()
		if err != nil {
			c.Error("Error installing frontend", err)
		}

		// Configure CloudNativePG
		err = c.ConfigureBackend()
		if err != nil {
			c.Error("Error installing backend", err)
		}

		// Run Django migrations
		err = c.InitBackend()
		if err != nil {
			c.Error("Error running migrations to initialize backend", err)
		}

		// Deploy prometheus
		err = c.ConfigureMonitoring()
		if err != nil {
			c.Error("Error installing monitoring", err)
		}

		// Create job that creates superuser
		err = c.CreateAdminUser()
		if err != nil {
			c.Error("Error creating superuser", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

}
