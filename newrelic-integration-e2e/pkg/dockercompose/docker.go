package dockercompose

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/runtime/logger"
)

const (
	dockerComposeBin = "docker-compose"
	dockerBin        = "docker"
)

type Compose struct {
	path      string
	env       map[string]string
	buildArgs map[string]string

	cmdLogger logger.CommandLogger
}

// New creates a new docker-compose wrapper given the path to a docker-compose.yml file.
func New(filepath string) *Compose {
	return &Compose{
		path:      filepath,
		cmdLogger: logger.NewGHALogger(os.Stderr),
	}
}

func (c *Compose) Logger(commandLogger logger.CommandLogger) {
	c.cmdLogger = commandLogger
}

func (c *Compose) Env(env map[string]string) {
	c.env = env
}

func (c *Compose) BuildArgs(args map[string]string) {
	c.buildArgs = args
}

func (c *Compose) command(args ...string) error {
	cmdArgs := []string{"-f", c.path}
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(dockerComposeBin, cmdArgs...)

	logWriter := c.cmdLogger.Open(cmd.String())
	defer c.cmdLogger.Close()

	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	return cmd.Run()
}

func (c *Compose) Run(container string) error {
	if err := c.Build(); err != nil {
		return fmt.Errorf("building before run: %w", err)
	}

	args := []string{"run", "-d", container}
	args = append(args, asFlags("-e", c.env)...)

	return c.command(args...)
}

func (c *Compose) Down() error {
	return c.command("down", "-v")
}

func (c *Compose) Build() error {
	args := []string{"build", "--no-cache"}
	// docker-compose.yml might use env vars to populate build args, so we set them during Build() as well.
	args = append(args, asFlags("--build-arg", c.env)...)
	args = append(args, asFlags("--build-arg", c.buildArgs)...)

	return c.command(args...)
}

func (c *Compose) Logs(containerName string) error {
	containerID := getContainerID(c.path, containerName)

	args := []string{"logs", containerID}
	cmd := exec.Command(dockerBin, args...)

	logWriter := c.cmdLogger.Open(cmd.String())
	defer c.cmdLogger.Close()

	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	return cmd.Run()
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

func asFlags(flag string, vars map[string]string) (args []string) {
	for k, v := range vars {
		args = append(args, flag, fmt.Sprintf("%s=%s", k, v))
	}

	return
}
