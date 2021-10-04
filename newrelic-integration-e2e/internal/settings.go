package e2e

import (
	"io/ioutil"
	"path/filepath"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/sirupsen/logrus"
)

const customTagKey = "testKey"

var defaultSettingsOptions = settingOptions{
	logLevel: logrus.InfoLevel,
}

type settingOptions struct {
	logLevel      logrus.Level
	specPath      string
	specParentDir string
	licenseKey    string
	agentDir      string
	rootDir       string
	accountID     int
	apiKey        string
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

type Settings interface {
	Logger() *logrus.Logger
	SpecDefinition() *spec.Definition
	AgentDir() string
	RootDir() string
	SpecParentDir() string
	LicenseKey() string
	ApiKey() string
	AccountID() int
	CustomTagKey() string
}

type settings struct {
	logger         *logrus.Logger
	specDefinition *spec.Definition
	specParentDir  string
	rootDir        string
	agentDir       string
	licenseKey     string
	accountID      int
	apiKey         string
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

func (s *settings) CustomTagKey() string {
	return customTagKey
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
		specParentDir:  options.specParentDir,
		rootDir:        options.rootDir,
		licenseKey:     options.licenseKey,
		apiKey:         options.apiKey,
		accountID:      options.accountID,
	}, nil
}
