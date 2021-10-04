package agent

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"path/filepath"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/dockercompose"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/oshelper"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

const (
	integrationsCfgDir = "integrations.d"
	exportersDir       = "exporters"
	integrationsBinDir = "bin"
	dockerCompose      = "docker-compose.yml"
	defConfigFile      = "nri-config.yml"
	containerName      = "agent"
	infraAgentDir      = "newrelic-infra-agent"
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
	exportersDir      string
	binsDir           string
	licenseKey        string
	defConfigFile     string
	specParentDir     string
	dockerComposePath string
	logger            *logrus.Logger
	overrides         *spec.Agent
	customTagKey      string
}

func NewAgent(settings e2e.Settings) *agent {
	agentDir := settings.AgentDir()

	return &agent{
		specParentDir:     settings.SpecParentDir(),
		agentDir:          agentDir,
		configsDir:        filepath.Join(agentDir, infraAgentDir, integrationsCfgDir),
		exportersDir:      filepath.Join(agentDir, infraAgentDir, exportersDir),
		binsDir:           filepath.Join(agentDir, infraAgentDir, integrationsBinDir),
		defConfigFile:     filepath.Join(agentDir, infraAgentDir, integrationsCfgDir, defConfigFile),
		dockerComposePath: filepath.Join(agentDir, dockerCompose),
		licenseKey:        settings.LicenseKey(),
		logger:            settings.Logger(),
		overrides:         settings.SpecDefinition().AgentOverrides,
		customTagKey:      settings.CustomTagKey(),
	}
}

func (a *agent) initialize() error {
	a.logger.Debug("removing temporary folders")
	if err := oshelper.RemoveDirectories(a.exportersDir, a.configsDir, a.binsDir); err != nil {
		return err
	}
	a.logger.Debug("creating folders required by the agent")
	return oshelper.MakeDirs(0777, a.exportersDir, a.configsDir, a.binsDir)
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
	for k, v := range a.overrides.Integrations {
		source := filepath.Join(a.specParentDir, v)
		destination := filepath.Join(a.binsDir, k)
		return oshelper.CopyFile(source, destination)
	}
	return nil
}

func (a *agent) Run(scenarioTag string) error {
	return dockercompose.Run(a.dockerComposePath, containerName, map[string]string{
		"NRIA_VERBOSE":           "1",
		"NRIA_LICENSE_KEY":       a.licenseKey,
		"NRIA_CUSTOM_ATTRIBUTES": fmt.Sprintf(`{"%s":"%s"}`, a.customTagKey, scenarioTag),
	})
}

func (a *agent) Stop() error {
	if a.logger.GetLevel() == logrus.DebugLevel {
		a.logger.Debug(dockercompose.Logs(a.dockerComposePath, containerName))
	}
	return dockercompose.Down(a.dockerComposePath)
}
