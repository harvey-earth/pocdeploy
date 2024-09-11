package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/harvey-earth/pocdeploy/internal"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a kubernetes cluster and deploy",
	Long: `creates a Kind Kubernetes cluster and deploys a frontend application with a CloudNative PG backend.

If the type is not set, the default is a local Kind cluster.`,
	Example: `pocdeploy create -t [kind]`,
	Run: func(cmd *cobra.Command, args []string) {
		frontendType := "django"
		// Create cluster
		if viper.GetString("type") == "kind" {
			err := internal.CreateKindCluster(viper.GetString("name"))
			if err != nil {
				internal.Error("Error creating Kind cluster", err)
			}
		}

		// Create namespaces
		if err := internal.CreateNamespaces(); err != nil {
			internal.Error("Error creating namespaces", err)
		}

		// Install backend (CloudNativePG operator)
		if err := internal.InstallBackend(); err != nil {
			internal.Error("Error installing backend", err)
		}

		// Build image using name and version from config and return them
		imgName, imgVers, err := internal.BuildImage()
		if err != nil {
			internal.Error("Error building image", err)
		}

		// Install monitoring (prometheus operator)
		if err = internal.InstallMonitoring(); err != nil {
			internal.Error("Error installing monitoring", err)
		}

		// Load docker image
		if viper.GetString("type") == "kind" {
			if err := internal.LoadKindImage(imgName, imgVers); err != nil {
				internal.Error("Error loading image to Kind", err)
			}
		}

		// Deploy frontend with generated secret key
		if err = internal.CreateSecretKeySecret(); err != nil {
			internal.Error("Error creating namespaces", err)
		}

		if err = internal.ConfigureFrontend(); err != nil {
			internal.Error("Error installing frontend", err)
		}

		// Configure CloudNativePG
		if err = internal.ConfigureBackend(); err != nil {
			internal.Error("Error installing backend", err)
		}

		// Run Django migrations
		if err = internal.InitBackend(frontendType); err != nil {
			internal.Error("Error running migrations to initialize backend", err)
		}

		// Deploy prometheus
		if err = internal.ConfigureMonitoring(); err != nil {
			internal.Error("Error installing monitoring", err)
		}

		// Create job that creates superuser
		if err = internal.CreateAdminUser(); err != nil {
			internal.Error("Error creating superuser", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
