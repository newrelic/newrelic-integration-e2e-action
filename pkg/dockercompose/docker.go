package dockercompose

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	dockerBin = "docker"
)

func Run(path string, container string, envVars map[string]string) error {
	if err := Build(path, container, envVars); err != nil {
		return err
	}
	args := []string{"compose", "-f", path, "run"}
	for k, v := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, "-d", container)
	cmd := exec.Command(dockerBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Down(path string) error {
	args := []string{"compose", "-f", path, "down", "-v"}
	cmd := exec.Command(dockerBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Build(path, container string, envVars map[string]string) error {
	args := []string{"compose", "-f", path, "build", "--no-cache"}
	for k, v := range envVars {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, container)
	cmd := exec.Command(dockerBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Logs(path, containerName string) string {
	containerID := getContainerID(path, containerName)

	args := []string{"logs", containerID}
	cmd := exec.Command(dockerBin, args...)
	stdout, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		fmt.Print(string(ee.Stderr))
	}
	return string(stdout)
}

func getContainerID(path, containerName string) string {
	const shortContainerIDLength = 12
	args := []string{"compose", "-f", path, "ps", "-q", containerName}
	cmd := exec.Command(dockerBin, args...)
	containerID, _ := cmd.Output()
	if len(containerID) > shortContainerIDLength {
		return string(containerID)[:shortContainerIDLength]
	}
	return string(containerID)
}
