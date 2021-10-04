package agent

import (
	"github.com/newrelic/newrelic-integration-e2e-action/newrelic-integration-e2e/internal/spec"
)

type integration struct {
	Name   string                 `yaml:"name"`
	Config map[string]interface{} `yaml:"config"`
}
type integrationList struct {
	Integrations []integration `yaml:"integrations"`
}

func getIntegrationList(integrations []spec.Integration) *integrationList {
	out := &integrationList{
		Integrations: make([]integration, len(integrations)),
	}
	for index, in := range integrations {
		out.Integrations[index] = integration{
			Name:   in.Name,
			Config: in.Config,
		}
	}
	return out
}
