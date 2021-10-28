package e2e

import (
	"io/ioutil"
	"path/filepath"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
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
	agentDir      string
	agentEnabled  bool
	rootDir       string
	accountID     int
	apiKey        string
	retryAttempts int
	retrySeconds  int
	commitSha     string
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

func SettingsWithAgentDir(agentDir string) SettingOption {
	return func(o *settingOptions) {
		o.agentDir = agentDir
	}
}

func SettingsWithRootDir(rootDir string) SettingOption {
	return func(o *settingOptions) {
		o.rootDir = rootDir
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

type Settings interface {
	Logger() *logrus.Logger
	SpecDefinition() *spec.Definition
	AgentEnabled() bool
	AgentDir() string
	RootDir() string
	SpecParentDir() string
	LicenseKey() string
	ApiKey() string
	AccountID() int
	RetryAttempts() int
	RetrySeconds() int
	CommitSha() string
}

type settings struct {
	logger         *logrus.Logger
	specDefinition *spec.Definition
	agentEnabled   bool
	specParentDir  string
	rootDir        string
	agentDir       string
	licenseKey     string
	accountID      int
	apiKey         string
	retryAttempts  int
	retrySeconds   int
	commitSha      string
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

func (s *settings) AgentDir() string {
	return s.agentDir
}

func (s *settings) AgentEnabled() bool {
	return s.agentEnabled
}

func (s *settings) RootDir() string {
	return s.rootDir
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
		agentDir:       options.agentDir,
		agentEnabled:   options.agentEnabled,
		specParentDir:  options.specParentDir,
		rootDir:        options.rootDir,
		licenseKey:     options.licenseKey,
		apiKey:         options.apiKey,
		accountID:      options.accountID,
		retryAttempts:  options.retryAttempts,
		retrySeconds:   options.retrySeconds,
		commitSha:      options.commitSha,
	}, nil
}
