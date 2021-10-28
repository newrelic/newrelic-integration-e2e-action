package main

import (
	_ "embed"
	"flag"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/newrelic"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/runtime"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/sirupsen/logrus"
)

const (
	flagSpecPath      = "spec_path"
	flagVerboseMode   = "verbose_mode"
	flagApiKey        = "api_key"
	flagAccountID     = "account_id"
	flagLicenseKey    = "license_key"
	flagAgentDir      = "agent_dir"
	flagAgentEnabled  = "agent_enabled"
	flagRetryAttempts = "retry_attempts"
	flagRetrySecons   = "retry_seconds"
	flagCommitSha     = "commit_sha"
)

func processCliArgs() (string, string, string, bool, string, int, int, int, string, logrus.Level) {
	specsPath := flag.String(flagSpecPath, "", "Relative path to the spec file")
	licenseKey := flag.String(flagLicenseKey, "", "New Relic License Key")
	agentDir := flag.String(flagAgentDir, "", "Directory used to deploy the agent")
	agentEnabled := flag.Bool(flagAgentEnabled, true, "If false the agent is not run")
	verboseMode := flag.Bool(flagVerboseMode, false, "If true the debug level is enabled")
	apiKey := flag.String(flagApiKey, "", "New Relic Api Key")
	accountID := flag.Int(flagAccountID, 0, "New Relic accountID to be used")
	retryAttempts := flag.Int(flagRetryAttempts, 10, "Number of attempts to retry a test")
	retrySeconds := flag.Int(flagRetrySecons, 30, "Number of seconds before retrying a test")
	commitSha := flag.String(flagCommitSha, "", "Current commit sha")
	flag.Parse()

	if *licenseKey == "" {
		logrus.Fatalf("missing required license_key")
	}
	if *specsPath == "" {
		logrus.Fatalf("missing required spec_path")
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
	return *licenseKey, *specsPath, *agentDir, *agentEnabled, *apiKey, *accountID, *retryAttempts, *retrySeconds, *commitSha, logLevel

}

func main() {
	logrus.Info("running e2e")

	licenseKey, specsPath, agentDir, agentEnabled, apiKey, accountID, retryAttempts, retrySeconds, commitSha, logLevel := processCliArgs()
	s, err := e2e.NewSettings(
		e2e.SettingsWithSpecPath(specsPath),
		e2e.SettingsWithLogLevel(logLevel),
		e2e.SettingsWithLicenseKey(licenseKey),
		e2e.SettingsWithAgentEnabled(agentEnabled),
		e2e.SettingsWithAgentDir(agentDir),
		e2e.SettingsWithApiKey(apiKey),
		e2e.SettingsWithAccountID(accountID),
		e2e.SettingsWithRetryAttempts(retryAttempts),
		e2e.SettingsWithRetrySeconds(retrySeconds),
		e2e.SettingsWithCommitSha(commitSha),
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

	nrClient := newrelic.NewNrClient(settings.ApiKey(), settings.AccountID())

	runtimeTester := []runtime.Tester{
		runtime.NewEntitiesTester(nrClient, settings.Logger()),
		runtime.NewMetricsTester(nrClient, settings.Logger(), settings.SpecParentDir()),
		runtime.NewNRQLTester(nrClient, settings.Logger()),
	}

	return runtime.NewRunner(runtimeTester, settings), nil
}
