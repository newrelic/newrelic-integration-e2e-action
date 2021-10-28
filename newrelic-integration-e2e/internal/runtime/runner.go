package runtime

import (
	"math/rand"
	"os"
	"os/exec"
	"time"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/agent"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/retrier"
	"github.com/sirupsen/logrus"
)

const (
	dmTableName       = "Metric"
	scenarioTagRuneNr = 5
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

type Tester interface {
	Test(tests spec.Tests, customTagKey, customTagValue string) []error
}

type Runner struct {
	agent         agent.Agent
	agentEnabled  bool
	testers       []Tester
	logger        *logrus.Logger
	spec          *spec.Definition
	specParentDir string
	retryAttempts int
	retryAfter    time.Duration
	commitSha     string
}

func NewRunner(testers []Tester, settings e2e.Settings) *Runner {
	rand.Seed(time.Now().UnixNano())

	var retryAttempts int
	if settings.RetryAttempts() > 0 {
		retryAttempts = settings.RetryAttempts()
	}

	var retryAfter time.Duration
	if settings.RetrySeconds() > 0 {
		retryAfter = time.Duration(settings.RetrySeconds()) * time.Second
	}

	var agentInstance agent.Agent
	if settings.AgentEnabled() {
		agentInstance = agent.NewAgent(settings)
	}

	return &Runner{
		agent:         agentInstance,
		agentEnabled:  settings.AgentEnabled(),
		testers:       testers,
		logger:        settings.Logger(),
		spec:          settings.SpecDefinition(),
		specParentDir: settings.SpecParentDir(),
		retryAttempts: retryAttempts,
		retryAfter:    retryAfter,
		commitSha:     settings.CommitSha(),
	}
}

func (r *Runner) Run() error {
	for _, scenario := range r.spec.Scenarios {
		scenarioTag := r.generateScenarioTag()
		r.logger.Debugf("[scenario]: %s, [Tag]: %s", scenario.Description, scenarioTag)

		if err := r.executeOSCommands(scenario.Before, scenarioTag); err != nil {
			return err
		}

		if r.agentEnabled {
			if err := r.agent.SetUp(scenario); err != nil {
				return err
			}

			if err := r.agent.Run(scenarioTag); err != nil {
				return err
			}
		}

		errAssertions := r.executeTests(scenario.Tests, r.spec.CustomTestKey, scenarioTag)

		if err := r.executeOSCommands(scenario.After, scenarioTag); err != nil {
			r.logger.Error(err)
		}

		if r.agentEnabled {
			if err := r.agent.Stop(); err != nil {
				return err
			}
		}

		if errAssertions != nil {
			return errAssertions
		}
	}

	return nil
}

func (r *Runner) executeOSCommands(statements []string, scenarioTag string) error {
	for _, stmt := range statements {
		r.logger.Debugf("execute command '%s' from path '%s'", stmt, r.specParentDir)
		cmd := exec.Command("bash", "-c", stmt)
		cmd.Dir = r.specParentDir
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "SCENARIO_TAG="+scenarioTag)
		combinedOutput, err := cmd.CombinedOutput()
		r.logger.Debugf("stdout: %q", combinedOutput)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) executeTests(tests spec.Tests, customTestKey string, scenarioTag string) error {
	for _, tester := range r.testers {
		err := retrier.Retry(r.logger, r.retryAttempts, r.retryAfter, func() []error {
			return tester.Test(tests, customTestKey, scenarioTag)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) generateScenarioTag() string {
	b := make([]rune, scenarioTagRuneNr)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return r.commitSha + string(b)
}
