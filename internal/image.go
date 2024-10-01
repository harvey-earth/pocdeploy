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
	Info("Building docker image")

	// Set variables
	frontendType := viper.GetString("frontend.type")
	path := viper.GetString("frontend.path")
	patchDir := viper.GetString("frontend.patch_dir")
	dockerfile := (viper.GetString("frontend.dockerfile"))
	image := viper.GetString("frontend.image")
	vers = viper.GetString("frontend.version")
	imgStr := image + ":" + vers
	cmd := exec.Command("docker", "build", path, "-t", imgStr, "-f", dockerfile)

	// Patch using build/patches
	if err = cmdApplyPatches(path, patchDir); err != nil {
		err = fmt.Errorf("error applying patches: %w", err)
		return "", "", err
	}

	if frontendType == "django" {
		// Copy requirements.txt if none exists for Django
		if err = copyRequirements(path); err != nil {
			err = fmt.Errorf("error copying requirements.txt: %w", err)
			return "", "", err
		}
	}

	// Build image
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error building docker image: %w", err)
		return "", "", err
	}

	Info("Docker image " + imgStr + " built")
	return image, vers, nil
}

func cmdApplyPatches(repo string, patchDir string) error {
	// Check if patchDir is empty
	patchPath, err := filepath.Abs(patchDir)
	if err != nil {
		err = fmt.Errorf("error getting absolute path of patch directory: %w", err)
		return err
	}
	patchFiles, err := os.ReadDir(patchPath)
	if err != nil {
		err = fmt.Errorf("error reading patch files: %w", err)
		return err
	}
	if len(patchFiles) == 0 {
		msg := fmt.Sprintf("%s empty directory, skipping...\n", patchDir)
		Debug(msg)
		return nil
	}
	repoPath, err := filepath.Abs(repo)
	if err != nil {
		err = fmt.Errorf("error getting absolute path of frontend path: %w", err)
		return err
	}

	msg := fmt.Sprintf("Applying patches from %s to %s", patchDir, repo)
	Debug(msg)
	// Iterate through patch files
	for _, f := range patchFiles {
		filename := patchPath + "/" + f.Name()
		cmd := exec.Command("git", "apply", filename)
		cmd.Dir = repoPath

		if err := cmd.Run(); err != nil {
			err = fmt.Errorf("error applying patch from %s: %w", filename, err)
			return err
		}
	}

	Debug("Applied patches")
	return nil
}

func copyRequirements(dest string) error {
	dst := filepath.Join(dest, "requirements.txt")

	if _, err := os.Stat(dst); err == nil {
		// File exists, skip copying
		msg := fmt.Sprintf("File %s already exists, skipping...\n", dst)
		Debug(msg)
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred while checking file existence
		return err
	}

	Debug("Proceeding to copy in frontend-requirements.txt")
	// Open the source file
	srcFile, err := d.DeployFiles.Open("frontend/frontend-requirements.txt")
	if err != nil {
		err = fmt.Errorf("error opening embeded frontend-requirements.txt file: %w", err)
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		err = fmt.Errorf("error creating requirements.txt file: %w", err)
		return err
	}
	defer dstFile.Close()

	// Copy the file contents from source to destination
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		err = fmt.Errorf("error copying requirements.txt file: %w", err)
		return err
	}

	return nil
}
