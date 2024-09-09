package internal

import (
	_ "embed" // Needed to use DeployFiles
	"fmt"
	"os"
	"os/exec"

	d "github.com/harvey-earth/pocdeploy/deploy"
)

// CreateKindCluster runs a shell command to create a Kind cluster
func CreateKindCluster() error {
	fmt.Println("Creating kind cluster")

	// Get embedded config
	configContent, err := d.DeployFiles.ReadFile("kind/config/kind-config.yaml")
	if err != nil {
		return err
	}
	tempfile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	if _, err := tempfile.Write(configContent); err != nil {
		return err
	}
	if err = tempfile.Close(); err != nil {
		return err
	}
	cfgName := tempfile.Name()

	// Create cluster with config
	cmd := exec.Command("kind", "create", "cluster", "--config", cfgName)
	if err := cmd.Run(); err != nil {
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
