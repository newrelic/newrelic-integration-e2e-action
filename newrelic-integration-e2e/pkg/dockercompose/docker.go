package dockercompose

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	dockerComposeBin = "docker-compose"
	dockerBin        = "docker"
)

func Run(path string, container string, envVars map[string]string) error {
	if err := Build(path, container); err != nil {
		return err
	}
	args := []string{"-f", path, "run"}
	for k, v := range envVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, "-d", container)
	cmd := exec.Command(dockerComposeBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Down(path string) error {
	args := []string{"-f", path, "down", "-v"}
	cmd := exec.Command(dockerComposeBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Build(path, container string) error {
	args := []string{"-f", path, "build", "--no-cache", container}
	fmt.Println(strings.Join(args, " "))
	cmd := exec.Command(dockerComposeBin, args...)
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
	args := []string{"-f", path, "ps", "-q", containerName}
	cmd := exec.Command(dockerComposeBin, args...)
	containerID, _ := cmd.Output()
	if len(containerID) > shortContainerIDLength {
		return string(containerID)[:shortContainerIDLength]
	}
	return string(containerID)
}
