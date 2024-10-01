package cmd

import (
	"fmt"

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
		frontendType := viper.GetString("frontend.type")
		clusterType := viper.GetString("type")
		// Create cluster
		if clusterType == "kind" {
			err := internal.CreateKindCluster(viper.GetString("name"))
			if err != nil {
				err = fmt.Errorf("Error creating Kind cluster: %w", err)
				internal.Error(err)
			}
		}

		// Create namespaces
		if err := internal.CreateNamespaces(); err != nil {
			err = fmt.Errorf("Error creating namespaces: %w", err)
			internal.Error(err)
		}

		// Install backend (CloudNativePG operator)
		if err := internal.InstallBackend(); err != nil {
			err = fmt.Errorf("Error installing backend: %w", err)
			internal.Error(err)
		}

		// Build image using name and version from config and return them
		imgName, imgVers, err := internal.BuildImage()
		if err != nil {
			err = fmt.Errorf("Error building image: %w", err)
			internal.Error(err)
		}

		// Install monitoring (prometheus operator)
		if err = internal.InstallMonitoring(); err != nil {
			err = fmt.Errorf("Error installing monitoring: %w", err)
			internal.Error(err)
		}

		// Load docker image
		if clusterType == "kind" {
			if err := internal.LoadKindImage(imgName, imgVers); err != nil {
				err = fmt.Errorf("Error loading image to Kind: %w", err)
				internal.Error(err)
			}
		}

		// Deploy frontend with generated secret key
		if err = internal.CreateSecretKeySecret(); err != nil {
			err = fmt.Errorf("Error creating namespace: %w", err)
			internal.Error(err)
		}

		if err = internal.ConfigureFrontend(); err != nil {
			err = fmt.Errorf("Error installing frontend: %w", err)
			internal.Error(err)
		}

		// Configure CloudNativePG
		if err = internal.ConfigureBackend(); err != nil {
			err = fmt.Errorf("Error installing backend: %w", err)
			internal.Error(err)
		}

		// Run Django migrations
		if err = internal.InitBackend(frontendType); err != nil {
			err = fmt.Errorf("Error running migrations to init backend: %w", err)
			internal.Error(err)
		}

		// Deploy prometheus
		if err = internal.ConfigureMonitoring(); err != nil {
			err = fmt.Errorf("Error installing monitoring: %w", err)
			internal.Error(err)
		}

		if frontendType == "django" {
			// Create job that creates superuser for Django
			if err = internal.CreateDjangoAdminUser(); err != nil {
				err = fmt.Errorf("Error creating superuser: %w", err)
				internal.Error(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
