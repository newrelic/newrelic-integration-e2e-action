package agent

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/dockercompose"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/oshelper"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

const (
	integrationsCfgDir    = "integrations.d"
	integrationsCfgDirEnv = "E2E_NRI_CONFIG"
	exportersDir          = "exporters"
	exportersDirEnv       = "E2E_EXPORTER_BIN"
	integrationsBinDir    = "bin"
	integrationsBinDirEnv = "E2E_NRI_BIN"
	dockerCompose         = "docker-compose.yml"
	defConfigFile         = "nri-config.yml"
	container             = "agent"
)

type Agent interface {
	SetUp(scenario spec.Scenario) error
	Run(scenarioTag string) error
	Stop() error
}

type agent struct {
	scenario          spec.Scenario
	agentDir          string
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
	agentDir := settings.AgentDir()

	a := agent{
		specParentDir:     settings.SpecParentDir(),
		containerName:     container,
		agentDir:          agentDir,
		dockerComposePath: filepath.Join(agentDir, dockerCompose),
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

func (a *agent) initialize() error {
	configDir, err := ioutil.TempDir(a.agentDir, integrationsCfgDir)
	if err != nil {
		return fmt.Errorf("creating configs dir: %w", err)
	}

	a.logger.Debugf("configs dir: %s", configDir)
	a.configsDir = configDir

	exportersDir, err := ioutil.TempDir(a.agentDir, exportersDir)
	if err != nil {
		return fmt.Errorf("creating exporters dir: %w", err)
	}

	a.logger.Debugf("exporters dir: %s", exportersDir)
	a.exportersDir = exportersDir

	binsDir, err := ioutil.TempDir(a.agentDir, integrationsBinDir)
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
// config files that are going to be mounted the agent.
func (a *agent) SetUp(scenario spec.Scenario) error {
	a.scenario = scenario
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

	if err := os.RemoveAll(a.binsDir); err != nil {
		return err
	}

	if err := os.RemoveAll(a.exportersDir); err != nil {
		return err
	}

	if err := os.RemoveAll(a.configsDir); err != nil {
		return err
	}

	return dockercompose.Down(a.dockerComposePath)
}
