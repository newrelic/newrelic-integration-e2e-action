package runtime

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/newrelic/newrelic-integration-e2e-action/internal/spec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type agentMock struct {
	SetupCalls  int
	RunCalls    int
	StopCalls   int
	ScenarioTag string
}

func (a *agentMock) SetUp(_ spec.Scenario) error {
	a.SetupCalls++
	return nil
}
func (a *agentMock) Run(scenarioTag string) error {
	a.RunCalls++
	a.ScenarioTag = scenarioTag
	return nil
}
func (a *agentMock) Stop() error {
	a.StopCalls++
	return nil
}

func TestRunner_Run(t *testing.T) {
	const commitSha = "1234567A-long-commit-sha"

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
		AgentExtensions: nil,
	}

	runner := Runner{
		agent:         &agentMock{},
		testers:       nil,
		logger:        log,
		spec:          &specDefinition,
		specParentDir: "parent-dir",
		commitSha:     commitSha,
	}

	err := runner.Run()
	require.NoError(t, err)

	require.Equal(t, 1, runner.agent.(*agentMock).SetupCalls)
	require.Equal(t, 1, runner.agent.(*agentMock).RunCalls)
	require.Equal(t, 1, runner.agent.(*agentMock).StopCalls)
	require.Contains(t, runner.agent.(*agentMock).ScenarioTag, "e2e-1234567-")
	require.Equal(t, 12+scenarioTagRuneNr, len(runner.agent.(*agentMock).ScenarioTag))
}

func TestRunner_RunWithTests(t *testing.T) {
	tests := []struct {
		name        string
		scripts     []string
		expectError bool
	}{
		{
			name:        "RunWithTests",
			scripts:     []string{"echo 'test script'"},
			expectError: false,
		},
		{
			name:        "RunWithTestsError",
			scripts:     []string{"echo 'test script'", "exit 1"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const commitSha = "1234567A-long-commit-sha"

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
						Tests: spec.Tests{
							NRQLs:    []spec.TestNRQL{},
							Entities: []spec.TestEntity{},
							Metrics:  []spec.TestMetrics{},
							Scripts:  tt.scripts,
						},
					},
				},
				AgentExtensions: nil,
			}

			runner := Runner{
				agent:         &agentMock{},
				testers:       nil,
				logger:        log,
				spec:          &specDefinition,
				specParentDir: "parent-dir",
				commitSha:     commitSha,
			}

			mkdirErr := os.Mkdir("parent-dir", 0755)
			require.NoError(t, mkdirErr)
			defer func() {
				err := os.RemoveAll("parent-dir")
				if err != nil {
					t.Fatal(err)
				}
			}()

			err := runner.Run()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, runner.agent.(*agentMock).SetupCalls)
				require.Equal(t, 1, runner.agent.(*agentMock).RunCalls)
				require.Equal(t, 1, runner.agent.(*agentMock).StopCalls)
			}

			require.Contains(t, runner.agent.(*agentMock).ScenarioTag, "e2e-1234567-")
			require.Equal(t, 12+scenarioTagRuneNr, len(runner.agent.(*agentMock).ScenarioTag))
		})
	}
}
