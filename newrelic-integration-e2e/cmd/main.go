package main

import (
	_ "embed"
	"flag"
	"fmt"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/agent"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/runtime"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/sirupsen/logrus"
)

const (
	flagSpecPath    = "spec_path"
	flagVerboseMode = "verbose_mode"
	flagApiKey      = "api_key"
	flagAccountID   = "account_id"
	flagLicenseKey  = "license_key"
	flagAgentDir    = "agent_dir"
	flagRootDir     = "root_dir"
)

func processCliArgs() (string, string, string, string, string, int, logrus.Level) {
	specsPath := flag.String(flagSpecPath, "", "Relative path to the spec file")
	licenseKey := flag.String(flagLicenseKey, "", "New Relic License Key")
	agentDir := flag.String(flagAgentDir, "", "Directory used to deploy the agent")
	rootDir := flag.String(flagRootDir, "", "workspace directory")
	verboseMode := flag.Bool(flagVerboseMode, false, "If true the debug level is enabled")
	apiKey := flag.String(flagApiKey, "", "New Relic Api Key")
	accountID := flag.Int(flagAccountID, 0, "New Relic accountID to be used")
	flag.Parse()

	if *licenseKey == "" {
		logrus.Fatalf("missing required license_key")
	}
	if *specsPath == "" {
		logrus.Fatalf("missing required spec_path")
	}
	if *rootDir == "" {
		logrus.Fatalf("missing required root_dir")
	}
	if *accountID == 0 {
		logrus.Fatalf("missing required accountID")
	}
	if *apiKey == "" {
		logrus.Fatalf("missing required apiKey")
	}

	logLevel := logrus.InfoLevel
	if *verboseMode {
		logLevel = logrus.DebugLevel
	}
	return *licenseKey, *specsPath, *rootDir, *agentDir, *apiKey, *accountID, logLevel

}

func main() {
	logrus.Info("running e2e")

	licenseKey, specsPath, rootDir, agentDir, apiKey, accountID, logLevel := processCliArgs()
	s, err := e2e.NewSettings(
		e2e.SettingsWithSpecPath(specsPath),
		e2e.SettingsWithLogLevel(logLevel),
		e2e.SettingsWithLicenseKey(licenseKey),
		e2e.SettingsWithAgentDir(agentDir),
		e2e.SettingsWithRootDir(rootDir),
		e2e.SettingsWithApiKey(apiKey),
		e2e.SettingsWithAccountID(accountID),
	)
	if err != nil {
		logrus.Fatalf("error loading settings: %s", err)
	}

	runner, err := createRunner(s)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := runner.Run(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("execution completed successfully!")
}

func createRunner(settings e2e.Settings) (*runtime.Runner, error) {
	settings.Logger().Debug("validating the spec definition")

	if err := settings.SpecDefinition().Validate(); err != nil {
		return nil, fmt.Errorf("error validating the spec definition: %s", err)
	}

	nrClient := newrelic.NewNrClient(settings.ApiKey(), settings.AccountID())
	entitiesTester := runtime.NewEntitiesTester(nrClient, settings.Logger())
	metricsTester := runtime.NewMetricsTester(nrClient, settings.Logger(), settings.SpecParentDir())
	nrqlTester := runtime.NewNRQLTester(nrClient, settings.Logger())

	return runtime.NewRunner(
		agent.NewAgent(settings),
		[]runtime.Tester{
			entitiesTester,
			metricsTester,
			nrqlTester,
		},
		settings,
	), nil
}
