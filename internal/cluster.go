package internal

import (
	_ "embed" // Needed to use DeployFiles
	"fmt"
	"os"
	"os/exec"
	"text/template"

	d "github.com/harvey-earth/pocdeploy/deploy"
	"github.com/harvey-earth/pocdeploy/internal/models"
	"github.com/spf13/viper"
)

// CreateKindCluster runs a shell command to create a Kind cluster
func CreateKindCluster(name string) error {
	fmt.Println("Creating kind cluster")
	cluster := models.KubernetesCluster{
		Name: name,
		Type: models.Kind,
		Port: viper.GetInt("port"),
	}

	// Get embedded config
	// configContent, err := d.DeployFiles.ReadFile("kind/config/kind-config.yaml")
	// if err != nil {
	// 	err = fmt.Errorf("error reading kind config: %w", err)
	// 	return err
	// }
	tempfile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		err = fmt.Errorf("error creating tempfile: %w", err)
		return err
	}
	defer os.Remove(tempfile.Name())

	tmpl, err := template.New("kind-config.yaml").ParseFS(d.DeployFiles, "kind/config/kind-config.tmpl.yaml")
	if err != nil {
		err = fmt.Errorf("error parsing kind-config template: %w", err)
		return err
	}
	err = tmpl.Execute(tempfile, cluster)
	// if _, err := tempfile.Write(configContent); err != nil {
	// 	err = fmt.Errorf("error writing tempfile: %w", err)
	// 	return err
	// }
	if err = tempfile.Close(); err != nil {
		err = fmt.Errorf("error closing tempfile: %w", err)
		return err
	}
	cfgName := tempfile.Name()

	// Create cluster with config
	cmd := exec.Command("kind", "create", "cluster", "--config", cfgName)
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error creating kind cluster: %w", err)
		return err
	}
	fmt.Println("Cluster created")
	return nil
}

// DeleteKindCluster runs a shell command to delete a Kind cluster
func DeleteKindCluster(name string) error {
	fmt.Println("Deleting kind cluster")
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error deleting kind cluster: %w", err)
		return err
	}
	fmt.Println("Cluster " + name + " deleted")
	return nil
}

// LoadKindImage loads a docker image to the Kind cluster
func LoadKindImage(name string, vers string) error {
	fmt.Println("Loading docker image to kind")

	img := name + ":" + vers
	cmd := exec.Command("kind", "load", "docker-image", img, "--name", "test")

	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println("Image loaded")
	return nil
}
