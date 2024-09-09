package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/viper"

	d "github.com/harvey-earth/pocdeploy/deploy"
)

// BuildImage builds a docker image using the set variables from pocdeploy.yaml and returns a name and version of the built image
func BuildImage() (name string, vers string, err error) {
	fmt.Println("Building docker image")

	// Set variables
	path := viper.GetString("frontend.path")
	patchDir := viper.GetString("frontend.patch_dir")
	dockerfile := (viper.GetString("frontend.dockerfile"))
	image := viper.GetString("frontend.image")
	vers = viper.GetString("frontend.version")
	imgStr := image + ":" + vers
	cmd := exec.Command("docker", "build", path, "-t", imgStr, "-f", dockerfile)

	// Patch using build/patches
	err = cmdApplyPatches(path, patchDir)
	if err != nil {
		return "", "", err
	}

	// Copy requirements.txt if none exists
	err = copyRequirements(path)
	if err != nil {
		return "", "", err
	}

	// Build image
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	fmt.Println("Docker image " + imgStr + " built")
	return image, vers, nil
}

func cmdApplyPatches(repo string, patchDir string) error {
	// Get absolute paths
	patchPath, err := filepath.Abs(patchDir)
	if err != nil {
		return err
	}
	repoPath, err := filepath.Abs(repo)
	if err != nil {
		return err
	}
	patchFiles, err := os.ReadDir(patchPath)
	if err != nil {
		return err
	}

	fmt.Printf("Applying patches from %s to %s\n", patchPath, repoPath)
	// Iterate through patch files
	for _, f := range patchFiles {
		filename := patchPath + "/" + f.Name()
		cmd := exec.Command("git", "apply", filename)
		cmd.Dir = repoPath

		if err := cmd.Run(); err != nil {
			return err
		}
	}
	fmt.Println("Applied patches")
	return nil
}

func copyRequirements(dest string) error {
	dst := filepath.Join(dest, "requirements.txt")

	if _, err := os.Stat(dst); err == nil {
		// File exists, skip copying
		fmt.Printf("File %s already exists, skipping...\n", dst)
		return err
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking file existence
		return err
	}

	fmt.Println("Proceeding to copy in frontend-requirements.txt")
	// Open the source file
	srcFile, err := d.DeployFiles.Open("frontend/frontend-requirements.txt")
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the file contents from source to destination
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
