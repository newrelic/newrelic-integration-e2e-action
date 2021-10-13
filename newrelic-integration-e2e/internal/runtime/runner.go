package runtime

import (
	"fmt"
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
	dmTableName       = "Metric"
	scenarioTagRuneNr = 10
	e2eDockerNetwork  = "e2e"
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
	retryAttempts int
	retryAfter    time.Duration
	commitSha     string
}

func NewRunner(agent agent.Agent, testers []Tester, settings e2e.Settings) *Runner {
	rand.Seed(time.Now().UnixNano())

	var retryAttempts int
	if settings.RetryAttempts() > 0 {
		retryAttempts = settings.RetryAttempts()
	}

	var retryAfter time.Duration
	if settings.RetrySeconds() > 0 {
		retryAfter = time.Duration(settings.RetrySeconds()) * time.Second
	}

	return &Runner{
		agent:         agent,
		testers:       testers,
		logger:        settings.Logger(),
		spec:          settings.SpecDefinition(),
		specParentDir: settings.SpecParentDir(),
		retryAttempts: retryAttempts,
		retryAfter:    retryAfter,
		customTagKey:  settings.CustomTagKey(),
		commitSha:     settings.CommitSha(),
	}
}

func (r *Runner) Run() error {
	r.createDockerE2ENetwork()
	defer r.removeDockerE2ENetwork()

	for _, scenario := range r.spec.Scenarios {
		scenarioTag := r.generateScenarioTag()
		r.logger.Debugf("[scenario]: %s, [Tag]: %s", scenario.Description, scenarioTag)

		if err := r.executeOSCommands(scenario.Before); err != nil {
			return err
		}

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
		err := retrier.Retry(r.logger, r.retryAttempts, r.retryAfter, func() []error {
			return tester.Test(tests, r.customTagKey, scenarioTag)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) createDockerE2ENetwork() {
	r.logger.Debugf("creating docker e2e network")
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker network create %s", e2eDockerNetwork))
	cmd.Dir = r.specParentDir
	stdout, _ := cmd.Output()
	logrus.Debug(stdout)
}

func (r *Runner) removeDockerE2ENetwork() {
	r.logger.Debugf("removing docker e2e network")
	cmd := exec.Command("bash", "-c", fmt.Sprintf("docker network remove %s", e2eDockerNetwork))
	cmd.Dir = r.specParentDir
	stdout, _ := cmd.Output()
	logrus.Debug(stdout)
}

func (r *Runner) generateScenarioTag() string {
	b := make([]rune, scenarioTagRuneNr)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return r.commitSha + string(b)
}
