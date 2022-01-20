package agent

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	e2e "github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal"
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/pkg/oshelper"
	"github.com/stretchr/testify/require"
)

func TestAgent_SetUp(t *testing.T) {
	agentDir := t.TempDir()
	rootDir := t.TempDir()

	_, err := os.Create(filepath.Join(rootDir, "/nri-powerdns"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(rootDir, "/nri-powerdns-exporter"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(rootDir, "/nri-prometheus"))
	require.NoError(t, err)
	err = oshelper.CopyFile("testdata/spec_file.yml", filepath.Join(rootDir, "spec_file.yml"))
	require.NoError(t, err)

	settings, err := e2e.NewSettings(
		e2e.SettingsWithSpecPath(filepath.Join(rootDir, "spec_file.yml")),
		e2e.SettingsWithAgentDir(agentDir),
	)
	require.NoError(t, err)

	t.Run("Given a scenario with 1 integration, the correct files should be in the AgentDir", func(t *testing.T) {
		sut := NewAgent(settings)
		require.NotEmpty(t, sut)

		err := sut.SetUp(settings.SpecDefinition().Scenarios[0])
		require.NoError(t, err)

		// nri-integration and exporter
		err = filepath.WalkDir(agentDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			files, err := ioutil.ReadDir(path)
			require.NoError(t, err)

			switch {
			case strings.Contains(d.Name(), integrationsBinDir):
				require.Equal(t, 2, len(files))

			case strings.Contains(d.Name(), exportersDir):
				require.Equal(t, 1, len(files))

			case strings.Contains(d.Name(), integrationsCfgDir):
				require.Equal(t, 1, len(files))

			case path == agentDir:
				return nil

			default:
				require.Fail(t, "found not expected directory", path)
			}

			return nil
		})

		require.NoError(t, err)
	})
}
