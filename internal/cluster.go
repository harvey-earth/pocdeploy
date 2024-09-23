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
	Info("Creating Kind cluster")

	clusterSize := make([]int, viper.GetInt("workers")-1)
	cluster := models.KubernetesCluster{
		Name: name,
		Type: models.Kind,
		Size: clusterSize,
	}

	tempfile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		err = fmt.Errorf("error creating tempfile: %w", err)
		return err
	}
	defer os.Remove(tempfile.Name())

	tmpl, err := template.New("kind-config.yaml.tmpl").ParseFS(d.DeployFiles, "kind/config/kind-config.yaml.tmpl")
	if err != nil {
		err = fmt.Errorf("error parsing kind-config template: %w", err)
		return err
	}
	if err = tmpl.Execute(tempfile, cluster); err != nil {
		err = fmt.Errorf("error executing template: %w", err)
		return err
	}
	if err = tempfile.Close(); err != nil {
		err = fmt.Errorf("error closing tempfile: %w", err)
		return err
	}

	// Create cluster with config
	cmd := exec.Command("kind", "create", "cluster", "--config", tempfile.Name())
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error creating kind cluster on command line: %w", err)
		return err
	}

	Info("Cluster created")
	return nil
}

// DeleteKindCluster runs a shell command to delete a Kind cluster
func DeleteKindCluster(name string) error {
	Info("Deleting Kind cluster")

	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error deleting kind cluster %s: %w", name, err)
		return err
	}

	Info("Cluster " + name + " deleted")
	return nil
}

// LoadKindImage loads a docker image to the Kind cluster
func LoadKindImage(name string, vers string) error {
	Debug("Loading docker image to Kind cluster")

	img := name + ":" + vers
	cmd := exec.Command("kind", "load", "docker-image", img, "--name", "test")

	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running kind load command: %w", err)
		return err
	}

	Debug("Image loaded to Kind cluster")
	return nil
}
