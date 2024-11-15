package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/v2/pexec"
)

func GenerateBuilder(jamPath string, rootDir string, outputDir string, registryUrl string) (buildImageUrl string, runImageUrl string, builderImageUrl string, err error) {

	buildArchive := filepath.Join(rootDir, outputDir, "build.oci")
	buildImageID := fmt.Sprintf("build-image-%s", uuid.NewString())
	buildImageUrl, err = PushFileToLocalRegistry(jamPath, buildArchive, registryUrl, buildImageID)
	if err != nil {
		return "", "", "", err
	}

	runArchive := filepath.Join(rootDir, outputDir, "run.oci")
	runImageID := fmt.Sprintf("run-image-%s", uuid.NewString())
	runImageUrl, err = PushFileToLocalRegistry(jamPath, runArchive, registryUrl, runImageID)
	if err != nil {
		return "", "", "", err
	}

	// Creating builder file
	builderConfigFile, err := os.CreateTemp("", "builder.toml")
	if err != nil {
		return "", "", "", err
	}

	builderConfigFilepath := builderConfigFile.Name()

	_, err = fmt.Fprintf(builderConfigFile, `
			[stack]
			  id = "io.buildpacks.stacks.ubi8"
			  build-image = "%s:latest"
			  run-image = "%s:latest"
			`, buildImageUrl, runImageUrl)

	if err != nil {
		return "", "", "", err
	}

	// naming builder and pushing it to registry with pack cli
	builderImageUrl = fmt.Sprintf("%s/builder-%s", registryUrl, uuid.NewString())

	buf := bytes.NewBuffer(nil)

	pack := pexec.NewExecutable("pack")
	err = pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"create",
			builderImageUrl,
			fmt.Sprintf("--config=%s", builderConfigFilepath),
			"--publish",
		},
	})

	if err != nil {
		return "", "", "", err
	}

	err = os.RemoveAll(builderConfigFilepath)
	if err != nil {
		return "", "", "", err
	}

	return buildImageUrl, runImageUrl, builderImageUrl, nil
}

func PushFileToLocalRegistry(jamPath string, filePath string, registryUrl string, imageName string) (string, error) {
	buf := bytes.NewBuffer(nil)

	imageURL := fmt.Sprintf("%s/%s", registryUrl, imageName)

	cmd := exec.Command(jamPath, "publish-image", "--image-ref", imageURL, "--image-archive", filePath)
	err := cmd.Run()

	if err != nil {
		return buf.String(), err
	} else {
		return imageURL, nil
	}
}

func RemoveImages(docker occam.Docker, imageIDs []string) error {

	for _, imageID := range imageIDs {
		err := docker.Image.Remove.Execute(imageID)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetLifecycleImageID(docker occam.Docker, builderImageUrl string) (lifecycleImageID string, err error) {

	lifecycleVersion, err := GetLifecycleVersion(builderImageUrl)
	if err != nil {
		return "", err
	}

	lifecycleImageID = fmt.Sprintf("buildpacksio/lifecycle:%s", lifecycleVersion)

	return lifecycleImageID, nil
}

type Builder struct {
	LocalInfo struct {
		Lifecycle struct {
			Version string `json:"version"`
		} `json:"lifecycle"`
	} `json:"remote_info"`
}

func GetLifecycleVersion(builderUrl string) (string, error) {
	buf := bytes.NewBuffer(nil)
	pack := pexec.NewExecutable("pack")
	err := pack.Execute(pexec.Execution{
		Stdout: buf,
		Stderr: buf,
		Args: []string{
			"builder",
			"inspect",
			builderUrl,
			"-o",
			"json",
		},
	})

	if err != nil {
		return "", err
	}

	var builder Builder
	err = json.Unmarshal([]byte(buf.String()), &builder)
	if err != nil {
		return "", err
	}
	return builder.LocalInfo.Lifecycle.Version, nil
}
