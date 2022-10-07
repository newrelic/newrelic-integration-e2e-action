package e2e

import (
	"io/ioutil"
	"path/filepath"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/sirupsen/logrus"
)

var defaultSettingsOptions = settingOptions{
	logLevel: logrus.InfoLevel,
}

type settingOptions struct {
	logLevel      logrus.Level
	specPath      string
	specParentDir string
	licenseKey    string
	agentEnabled  bool
	accountID     int
	apiKey        string
	retryAttempts int
	retrySeconds  int
	commitSha     string
	region        string
}

type SettingOption func(*settingOptions)

func SettingsWithSpecPath(specPath string) SettingOption {
	return func(o *settingOptions) {
		o.specPath = specPath
		o.specParentDir = filepath.Dir(specPath)
	}
}

func SettingsWithLogLevel(logLevel logrus.Level) SettingOption {
	return func(o *settingOptions) {
		o.logLevel = logLevel
	}
}

func SettingsWithLicenseKey(licenseKey string) SettingOption {
	return func(o *settingOptions) {
		o.licenseKey = licenseKey
	}
}

func SettingsWithAccountID(accountID int) SettingOption {
	return func(o *settingOptions) {
		o.accountID = accountID
	}
}

func SettingsWithApiKey(apiKey string) SettingOption {
	return func(o *settingOptions) {
		o.apiKey = apiKey
	}
}

func SettingsWithRetryAttempts(retryAttempts int) SettingOption {
	return func(o *settingOptions) {
		o.retryAttempts = retryAttempts
	}
}

func SettingsWithRetrySeconds(retrySeconds int) SettingOption {
	return func(o *settingOptions) {
		o.retrySeconds = retrySeconds
	}
}

func SettingsWithCommitSha(commitSha string) SettingOption {
	return func(o *settingOptions) {
		o.commitSha = commitSha
	}
}

func SettingsWithAgentEnabled(agentEnabled bool) SettingOption {
	return func(o *settingOptions) {
		o.agentEnabled = agentEnabled
	}
}

func SettingsWithRegion(region string) SettingOption {
	return func(o *settingOptions) {
		o.region = region
	}
}

type Settings interface {
	Logger() *logrus.Logger
	SpecDefinition() *spec.Definition
	AgentEnabled() bool
	AgentBuildContext() string
	SpecParentDir() string
	LicenseKey() string
	ApiKey() string
	AccountID() int
	RetryAttempts() int
	RetrySeconds() int
	CommitSha() string
	Region() string
}

type settings struct {
	logger         *logrus.Logger
	specDefinition *spec.Definition
	agentEnabled   bool
	specParentDir  string
	licenseKey     string
	accountID      int
	apiKey         string
	retryAttempts  int
	retrySeconds   int
	commitSha      string
	region         string
}

func (s *settings) Logger() *logrus.Logger {
	return s.logger
}

func (s *settings) LicenseKey() string {
	return s.licenseKey
}

func (s *settings) SpecDefinition() *spec.Definition {
	return s.specDefinition
}

func (s *settings) AgentBuildContext() string {
	if s.specDefinition == nil || s.specDefinition.AgentExtensions == nil {
		return ""
	}

	return filepath.Join(s.specParentDir, s.specDefinition.AgentExtensions.BuildContext)
}

func (s *settings) AgentEnabled() bool {
	return s.agentEnabled
}

func (s *settings) SpecParentDir() string {
	return s.specParentDir
}

func (s *settings) ApiKey() string {
	return s.apiKey
}

func (s *settings) AccountID() int {
	return s.accountID
}

func (s *settings) RetryAttempts() int {
	return s.retryAttempts
}

func (s *settings) RetrySeconds() int {
	return s.retrySeconds
}

func (s *settings) CommitSha() string {
	return s.commitSha
}

func (s *settings) Region() string {
	if s.region != "" {
		return s.region
	}
	return "US"
}

// New returns a Scheduler
func NewSettings(
	opts ...SettingOption) (Settings, error) {
	options := defaultSettingsOptions
	for _, opt := range opts {
		opt(&options)
	}
	logger := logrus.New()
	logger.SetLevel(options.logLevel)
	content, err := ioutil.ReadFile(options.specPath)
	if err != nil {
		return nil, err
	}
	logger.Debug("parsing the content of the spec file")
	s, err := spec.ParseDefinitionFile(content)
	if err != nil {
		return nil, err
	}
	logger.Debug("return with settings")

	return &settings{
		logger:         logger,
		specDefinition: s,
		agentEnabled:   options.agentEnabled,
		specParentDir:  options.specParentDir,
		licenseKey:     options.licenseKey,
		apiKey:         options.apiKey,
		accountID:      options.accountID,
		retryAttempts:  options.retryAttempts,
		retrySeconds:   options.retrySeconds,
		commitSha:      options.commitSha,
		region:         options.region,
	}, nil
}
