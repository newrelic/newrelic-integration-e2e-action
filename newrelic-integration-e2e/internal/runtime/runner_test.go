package runtime

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"

	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
)

type agentMock struct {
	SetupCalls   int
	RunCalls     int
	StopCalls    int
	ScenrarioTag string
}

func (a *agentMock) SetUp(_ spec.Scenario) error {
	a.SetupCalls++
	return nil
}
func (a *agentMock) Run(scenarioTag string) error {
	a.RunCalls++
	a.ScenrarioTag = scenarioTag
	return nil
}
func (a *agentMock) Stop() error {
	a.StopCalls++
	return nil
}

func TestRunner_Run(t *testing.T) {
	const commitSha = "a-commit-sha"

	log := logrus.New()
	log.SetOutput(ioutil.Discard)

	specDefinition := spec.Definition{
		Description: "definition",
		Scenarios: []spec.Scenario{
			{
				Description:  "empty-scenario",
				Integrations: nil,
				Before:       nil,
				After:        nil,
				Tests:        spec.Tests{},
			},
		},
		AgentOverrides: nil,
	}

	runner := Runner{
		agent:         &agentMock{},
		testers:       nil,
		logger:        log,
		spec:          &specDefinition,
		specParentDir: "parent-dir",
		customTagKey:  "tag-key",
		commitSha:     commitSha,
	}

	err := runner.Run()
	require.NoError(t, err)

	require.Equal(t, 1, runner.agent.(*agentMock).SetupCalls)
	require.Equal(t, 1, runner.agent.(*agentMock).RunCalls)
	require.Equal(t, 1, runner.agent.(*agentMock).StopCalls)
	require.Contains(t, runner.agent.(*agentMock).ScenrarioTag, commitSha)
	require.Equal(t, len(commitSha)+scenarioTagRuneNr, len(runner.agent.(*agentMock).ScenrarioTag))
}
