package runtime

import (
	"math/rand"
	"os/exec"
	"time"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/agent"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/retrier"
	"github.com/sirupsen/logrus"
)

const (
	dmTableName         = "Metric"
	retryNumberAttempts = 10
	retryAfter          = 30 * time.Second
	scenarioTagRuneNr   = 10
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

type Tester interface {
	Test(tests spec.Tests, customTagKey, customTagValue string) []error
}

type Runner struct {
	agent         agent.Agent
	testers       []Tester
	logger        *logrus.Logger
	spec          *spec.Definition
	specParentDir string
	customTagKey  string
}

func NewRunner(agent agent.Agent, testers []Tester, settings e2e.Settings) *Runner {
	rand.Seed(time.Now().UnixNano())

	return &Runner{
		agent:         agent,
		testers:       testers,
		logger:        settings.Logger(),
		spec:          settings.SpecDefinition(),
		specParentDir: settings.SpecParentDir(),
		customTagKey:  settings.CustomTagKey(),
	}
}

func (r *Runner) Run() error {
	for _, scenario := range r.spec.Scenarios {
		scenarioTag := r.generateScenarioTag()
		r.logger.Debugf("[scenario]: %s, [Tag]: %s", scenario.Description, scenarioTag)

		if err := r.agent.SetUp(scenario); err != nil {
			return err
		}

		if err := r.executeOSCommands(scenario.Before); err != nil {
			return err
		}

		if err := r.agent.Run(scenarioTag); err != nil {
			return err
		}

		errAssertions := r.executeTests(scenario.Tests, scenarioTag)

		if err := r.executeOSCommands(scenario.After); err != nil {
			r.logger.Error(err)
		}

		if err := r.agent.Stop(); err != nil {
			return err
		}

		if errAssertions != nil {
			return errAssertions
		}
	}

	return nil
}

func (r *Runner) executeOSCommands(statements []string) error {
	for _, stmt := range statements {
		r.logger.Debugf("execute command '%s' from path '%s'", stmt, r.specParentDir)
		cmd := exec.Command("bash", "-c", stmt)
		cmd.Dir = r.specParentDir
		stdout, err := cmd.Output()
		logrus.Debug(stdout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) executeTests(tests spec.Tests, scenarioTag string) error {
	for _, tester := range r.testers {
		err := retrier.Retry(r.logger, retryNumberAttempts, retryAfter, func() []error {
			return tester.Test(tests, r.customTagKey, scenarioTag)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO Improve tag with more info from each scenario, like GH commit
func (r *Runner) generateScenarioTag() string {
	b := make([]rune, scenarioTagRuneNr)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
