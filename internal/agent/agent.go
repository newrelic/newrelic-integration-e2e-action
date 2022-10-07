package agent

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/newrelic/newrelic-integration-e2e-action/pkg/dockercompose"
	"github.com/newrelic/newrelic-integration-e2e-action/pkg/oshelper"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

const (
	IntegrationsCfgDir    = "integrations.d"
	integrationsCfgDirEnv = "E2E_NRI_CONFIG"
	ExportersDir          = "exporters"
	exportersDirEnv       = "E2E_EXPORTER_BIN"
	IntegrationsBinDir    = "bin"
	integrationsBinDirEnv = "E2E_NRI_BIN"
	dockerCompose         = "docker-compose.yml"
	defConfigFile         = "nri-config.yml"
	container             = "agent"
)

//go:embed resources/docker-compose.yml
var defaultCompose []byte

type Agent interface {
	SetUp(scenario spec.Scenario) error
	Run(scenarioTag string) error
	Stop() error
}

type agent struct {
	agentBuildContext string
	configsDir        string
	containerName     string
	exportersDir      string
	binsDir           string
	licenseKey        string
	specParentDir     string
	dockerComposePath string
	logger            *logrus.Logger
	ExtraIntegrations map[string]string
	ExtraEnvVars      map[string]string
	customTagKey      string
}

func NewAgent(settings e2e.Settings) *agent {
	agentBuildContext := settings.AgentBuildContext()

	a := agent{
		specParentDir:     settings.SpecParentDir(),
		containerName:     container,
		agentBuildContext: agentBuildContext,
		dockerComposePath: filepath.Join(agentBuildContext, dockerCompose),
		licenseKey:        settings.LicenseKey(),
		logger:            settings.Logger(),
		customTagKey:      settings.SpecDefinition().CustomTestKey,
	}

	if settings.SpecDefinition().AgentExtensions != nil {
		a.ExtraIntegrations = settings.SpecDefinition().AgentExtensions.Integrations
		a.ExtraEnvVars = settings.SpecDefinition().AgentExtensions.EnvVars
	}

	return &a
}

// initDefaultCompose creates a temp dir with the embedded default docker-compose.yml .
func (a *agent) initDefaultCompose() error {
	if a.agentBuildContext != "" {
		return nil
	}

	composeDir, err := ioutil.TempDir("", "agent-docker-compose")
	if err != nil {
		return fmt.Errorf("crating default docker-compose dir: %w", err)
	}

	a.dockerComposePath = filepath.Join(composeDir, "docker-compose.yml")

	err = ioutil.WriteFile(
		a.dockerComposePath,
		defaultCompose,
		444,
	)
	if err != nil {
		return fmt.Errorf("crating default docker-compose file: %w", err)
	}

	a.logger.Debugf("using default docker-compose: %s", a.dockerComposePath)

	return nil
}

// initialize creates temp dirs for configs, and exporters/integrations bins inside
// the same dir where the agent compose file is located.
func (a *agent) initialize() error {
	// dockerComposePath can be in a temporal dir if using default or inside the
	// agentBuildContext dir if using custom agent container.
	parentDir := filepath.Dir(a.dockerComposePath)

	configDir, err := ioutil.TempDir(parentDir, IntegrationsCfgDir)
	if err != nil {
		return fmt.Errorf("creating configs dir: %w", err)
	}

	a.logger.Debugf("configs dir: %s", configDir)
	a.configsDir = configDir

	exportersDir, err := ioutil.TempDir(parentDir, ExportersDir)
	if err != nil {
		return fmt.Errorf("creating exporters dir: %w", err)
	}

	a.logger.Debugf("exporters dir: %s", exportersDir)
	a.exportersDir = exportersDir

	binsDir, err := ioutil.TempDir(parentDir, IntegrationsBinDir)
	if err != nil {
		return fmt.Errorf("creating integration bin dir: %w", err)
	}

	a.logger.Debugf("bins dir: %s", binsDir)
	a.binsDir = binsDir

	return nil
}

func (a *agent) addIntegration(integration spec.Integration) error {
	if integration.BinaryPath == "" {
		return nil
	}
	source := filepath.Join(a.specParentDir, integration.BinaryPath)
	destination := filepath.Join(a.binsDir, integration.Name)
	a.logger.Debugf("copy file from '%s' to '%s'", source, destination)
	return oshelper.CopyFile(source, destination)
}

func (a *agent) addPrometheusExporter(integration spec.Integration) error {
	if integration.ExporterBinaryPath == "" {
		return nil
	}
	exporterName := filepath.Base(integration.ExporterBinaryPath)
	source := filepath.Join(a.specParentDir, integration.ExporterBinaryPath)
	destination := filepath.Join(a.exportersDir, exporterName)
	a.logger.Debugf("copy file from '%s' to '%s'", source, destination)
	return oshelper.CopyFile(source, destination)
}

func (a *agent) addIntegrationsConfigFile(integrations []spec.Integration) error {
	content, err := yaml.Marshal(getIntegrationList(integrations))
	if err != nil {
		return err
	}
	cfgPath := filepath.Join(a.configsDir, defConfigFile)
	a.logger.Debugf("create config file '%s' in  '%s'", defConfigFile, cfgPath)
	return ioutil.WriteFile(cfgPath, content, 0777)
}

// SetUp creates temporary folders where it copies the binaries and
// config files that are going to be mounted in the agent container.
func (a *agent) SetUp(scenario spec.Scenario) error {
	if err := a.initDefaultCompose(); err != nil {
		return err
	}

	if err := a.initialize(); err != nil {
		return err
	}
	integrations := scenario.Integrations
	a.logger.Debugf("there are %d integrations", len(integrations))
	integrationsNames := make([]string, len(integrations))
	for i := range integrations {
		integration := integrations[i]
		if err := a.addIntegration(integration); err != nil {
			return err
		}
		if err := a.addPrometheusExporter(integration); err != nil {
			return err
		}
		integrationsNames[i] = integration.Name
	}
	if err := a.addIntegrationsConfigFile(integrations); err != nil {
		return err
	}
	for k, v := range a.ExtraIntegrations {
		source := filepath.Join(a.specParentDir, v)
		destination := filepath.Join(a.binsDir, k)
		return oshelper.CopyFile(source, destination)
	}
	return nil
}

func (a *agent) Run(scenarioTag string) error {
	envVars := map[string]string{
		"NRIA_VERBOSE":           "1",
		"NRIA_LICENSE_KEY":       a.licenseKey,
		"NRIA_CUSTOM_ATTRIBUTES": fmt.Sprintf(`{"%s":"%s"}`, a.customTagKey, scenarioTag),
	}

	for envKey, envValue := range a.ExtraEnvVars {
		envVars[envKey] = envValue
	}

	// Temporary directories with configs and binaries are passed to the docker-compose
	// through env vars. The docker compose is resposable for mounting this directories
	// so the Agent automatically executes the integrations.

	if err := os.Setenv(integrationsCfgDirEnv, a.configsDir); err != nil {
		return fmt.Errorf("fail to set %s env: %w", integrationsCfgDirEnv, err)
	}

	if err := os.Setenv(integrationsBinDirEnv, a.binsDir); err != nil {
		return fmt.Errorf("fail to set %s env: %w", integrationsBinDirEnv, err)
	}

	if err := os.Setenv(exportersDirEnv, a.exportersDir); err != nil {
		return fmt.Errorf("fail to set %s env: %w", exportersDirEnv, err)
	}

	return dockercompose.Run(a.dockerComposePath, a.containerName, envVars)
}

func (a *agent) Stop() error {
	if a.logger.GetLevel() == logrus.DebugLevel {
		a.logger.Debug(dockercompose.Logs(a.dockerComposePath, a.containerName))
	}

	if err := dockercompose.Down(a.dockerComposePath); err != nil {
		return err
	}

	// Remove compose file when using default.
	if a.agentBuildContext == "" {
		if err := os.RemoveAll(a.dockerComposePath); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(a.binsDir); err != nil {
		return err
	}

	if err := os.RemoveAll(a.exportersDir); err != nil {
		return err
	}

	if err := os.RemoveAll(a.configsDir); err != nil {
		return err
	}

	return nil
}
