package agent_test

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/agent"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/oshelper"
	"github.com/stretchr/testify/require"
)

func TestAgent_SetUp(t *testing.T) {
	specPath := t.TempDir()

	customBuildContext := filepath.Join(specPath, "build_context_dir")
	require.NoError(t, os.Mkdir(customBuildContext, fs.ModePerm))

	require.NoError(t, oshelper.CopyFile("testdata/spec_file.yml", filepath.Join(specPath, "spec_file.yml")))

	_, err := os.Create(filepath.Join(specPath, "/nri-powerdns"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(specPath, "/nri-powerdns-exporter"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(specPath, "/nri-prometheus"))
	require.NoError(t, err)

	settings, err := e2e.NewSettings(
		e2e.SettingsWithSpecPath(filepath.Join(specPath, "spec_file.yml")),
	)
	require.NoError(t, err)

	t.Run("Given a scenario with 1 integration, the correct files should be in the AgentDir", func(t *testing.T) {
		sut := agent.NewAgent(settings)
		require.NotEmpty(t, sut)

		err := sut.SetUp(settings.SpecDefinition().Scenarios[0])
		require.NoError(t, err)

		// nri-integration and exporter
		err = filepath.WalkDir(customBuildContext, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			files, err := ioutil.ReadDir(path)
			require.NoError(t, err)

			switch {
			case strings.Contains(d.Name(), agent.IntegrationsBinDir):
				require.Equal(t, 2, len(files))

			case strings.Contains(d.Name(), agent.ExportersDir):
				require.Equal(t, 1, len(files))

			case strings.Contains(d.Name(), agent.IntegrationsCfgDir):
				require.Equal(t, 1, len(files))

			case path == customBuildContext:
				return nil

			default:
				require.Fail(t, "found not expected directory", path)
			}

			return nil
		})

		require.NoError(t, err)
	})
}
