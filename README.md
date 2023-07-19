[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# newrelic-integration-e2e-action

End to end testing action for New Relic integrations to ensure that:

- The integrations are correctly executed by the agent.
- The service metrics and entities are correctly sent to NROne.

New Relic has two kinds of integrations:

- Custom made integrations
- Integrations based on prometheus exporters

![diagram](e2e.jpg)

## Steps executed by the e2e action

- It reads the e2e test descriptor file/s that must be passed as an argument to the action.
- For each scenario present in the descriptor:
  - It launches services dependencies (e.g. a docker-compose ) if specified in the `before` step.
  - It creates a config file with the details in the descriptor.
  - Adds a custom-attribute to the config:
    - Composed by the current commit sha + a new 10 alphanumeric-random digit on each scenario.
    - The tests will look for this label to fetch the metrics and the entities from the New Relic backend.
  - It launches the default docker-compose of the Infra Agent mounting the binaries and configs so the integrations are run automatically.
  - The runner executes the tests one by one, checking that metrics &/or entities are being created correctly.
  - If the test fails, it's retried after the `retry_seconds` (default 30s) and up to the `retry_attempts` (default 10) defined for the action.
  - It stops & removes the services if specified in the after step.
  - If `verbose` is true it logs the agent logs with other debug information.
- The action is completed.

## Install

`newrelic-integration-e2e-action` can be installed as binary by using the following Go tool command:

```shell
go install github.com/newrelic/newrelic-integration-e2e-action@latest
```

In case you don't want to install de binary you can always run it directly:

```shell
go run github.com/newrelic/newrelic-integration-e2e-action@latest
```

## Usage

Example usage:

```yaml
name: e2e
on:
  push:
    branches: ["powerdns_e2e"]
  workflow_dispatch:
    branches: ["powerdns_e2e"]
jobs:
  powerdns-e2e:
    name: E2E tests for PowerDNS
    runs-on: ubuntu-latest
    steps:
      - name: checkout-repository
        uses: actions/checkout@v2
      - name: Setup go
        uses: actions/setup-go@v1
        with:
          go-version: 1.18
      - name: build-powerdns
        run: make build-powerdns
      - name: e2e-test
        uses: newrelic/newrelic-integration-e2e-action@v1
        with:
          spec_path: exporters/powerdns/e2e/e2e_spec.yml
          account_id: ${{ secrets.ACCOUNT_ID }}
          api_key: ${{ secrets.API_KEY }}
          license_key: ${{ secrets.LICENSE_KEY }}
          retry_seconds: 30
          retry_attempts: 10
          verbose: false
```

The required fields are:

- `spec_path` to define the e2e.
- `account_id` required by the NR API.
- `api_key` required by the NR API (API key type: "User").
- `license_key` required by the agent (API key type: "Ingest - License").

Optional parameters:

- `retry_seconds` it's the number of seconds to wait after retrying a test. default: 30.
- `retry_attempts` it's the number of attempts a failed test can be retried. default: 10.
- `verbose` if set to to true the agent logs and other useful debug logs will be printed. default: false.
- `agent_enabled` if set to false then the agent will not be spawned and its lifecycle will be up to the user of the action. Useful when testing K8s like integrations
- `region` is where to send the e2e data. Possible values: "US", "EU", "Staging", "Local". See `action.yaml` for more info.
- `scenario_tag` is used as an environment variable in the spec file under `spec_path`. By default, the value of this variable is randomly generated. For now, our nri-kubernetes repo uses its random value as Kubernetes cluster and namespace names during the testing. Through this parameter, customers can set its value as their cluster name if they do not want to use random cluster name during the testing.

## Spec file for the e2e

The paths of the binaries in this file are relative to its parent folder.

The spec file for the e2e needs to be a yaml file with the following structure:

`decription` : Description for the e2e test.

`custom_test_key`: (Optional) Key of the custom attribute to test. Useful in case you cannot control the keyName.

`agent`: Extra environment variables and/or integrations required for the e2e.

- `build_context` : Relative path to the directory where a custom `docker-compose.yml` will be build and run to launch the Agent. If not specified a default embedded docker-compose is executed.
- `integrations` : Additional integrations needed for the e2e.
- `env_vars` : Additional EnvVars for the agent execution.

`scenarios`: Array of scenarios, each one is an independent run for the e2e.

- `decription` : Description of the scenario.
- `before` : Array of shell commands that will be executed by the e2e runner before the next steps of the scenario. (Here is where the docker-compose commands need to be put to setup the environment)
- `after` : Array of shell commands that will be executed by the e2e runner as the last step of the scenario.
- `integrations` : Array with the integrations running in this scenario.
  - `name` : Name of the integration under test.
  - `binary_path` : Relative path to the integration binary.
  - `exporter_binary_path` : Relative path to the prometheus exporter if it's needed (Prometheus based integrations)
  - `config` : The config values for this NR integration that will be red by the agent to execute the integration.
- `tests` : The 3 kinds of tests that will be done to the New relic api to check for metrics/entities in NROne:
  - `nrqls` : Array of queries that will be executed independently. You can specify if running a query an error is expected or not.
    - `query` : the query to run
    - `error_expected`: false by default, useful if we want to test that a metric is not being sent. This cannot be used in conjunction with `expected_results`.
    - `expected_results` : Array of expected results that will be sequentially asserted against the NRQL response.
      - `key`: The key of the expected result to assert against (i.e. `Pods Available`)
      - `value`: The value expected for the above key (i.e. `4`)
  - `metrics` : Array of metrics to check existing in NROne
    - `source` : Relative path to the integration spec file (It defines the entities and metrics) that will be parsed to match the metrics got from NROne.
    - `except_entities` : Array of entities whose metrics will be skipped.
    - `except_metrics` : Array of metrics to skip.
    - `exceptions_source` : Relative (to the spec file) path to a YAML file containing extra exceptions. This metrics are appended to the ones defined in `except_metrics` and `except_entities`.
  - `entities` : Array of entities to chek existing in NROne.
    - `type` : Type of the entity to look for in NROne
    - `data_type` : Name of the table to check for the entity in NROne (If V4 integration, will always be Metric)
    - `metric_name` : Name of the known metric that should be having the entity dimension in NROne.

Example:

```yaml
description: |
  End-to-end tests for PowerDNS integration

agent:
  env_vars:
    NRJMX_VERSION: "1.5.3"

scenarios:
  - description: |
      This scenario will verify that metrics froms PDNS authoritative
      are correcly collected.
    before:
      - docker-compose -f "deps/docker-compose.yml" up -d
    after:
      - docker-compose -f "deps/docker-compose.yml" down -d
    integrations:
      - name: nri-powerdns
        binary_path: bin/nri-powerdns
        exporter_binary_path: bin/nri-powerdns-exporter
        config:
          powerdns_url: http://localhost:8081/api/v1/
          exporter_port: 9121
          api_key: authoritative-secret
    tests:
      nrqls:
        - query: "SELECT average(powerdns_authoritative_queries_total) FROM Metric"
          error_expected: false
      entities:
        - type: "POWERDNS_AUTHORITATIVE"
          data_type: "Metric"
          metric_name: "powerdns_authoritative_up"
        - type: "POWERDNS_RECURSOR"
          data_type: "Metric"
          metric_name: "powerdns_recursor_up"
      metrics:
        - source: "powerdns.yml"
          except_entities:
            - POWERDNS_AUTHORITATIVE
          except_metrics:
            - powerdns_authoritative_answers_bytes_total
            - powerdns_recursor_cache_lookups_total
          exceptions_source: "powerdns-custom-exceptions.yml"
```

Extra exceptions file `powerdns-custom-exceptions.yml` example:

```yaml
except_entities:
  - POWERDNS_MY_CUSTOM_ENTITY
except_metrics:
  - powerdns_metric_removed_on_version_x

```

### Custom Agent image

A [docker-compose.yml](internal/agent/resources/docker-compose.yml) is embedded into the code which is used by default to build the Agent image that contains the integrations and configs to be tested.

If a custom image is needed, `agent.build_context` must contain a relative path to a directory containing a `docker-compose.yml`.In order to mount the binaries to the image `E2E_EXPORTER_BIN`, `E2E_NRI_CONFIG` and `E2E_NRI_BIN` will be set as env variables with the path to the temporary folders where assets are copied.

A concrete example can be checked in [this test](test/testdata/kafka/kafka-e2e.yml).

## Types of test

All the queries done to NROne are done with an extra WHERE condition that is `WHERE testKey = 'COMMMITSHA + 10 Digit alphanumeric'` a custom attribute added to the agent. This attribute is decorated in all the emitted metrics.

In this way we ensure that every returned metric/entity is really the emitted by the current e2e scenario.

Example:
`SELECT * from Metric where metricName = 'powerdns_authoritative_up' where testKey = '35e32b6a00dec02ae7d7c45c6b7106779a124685sneniedzku' limit 1`

### Entities

This test is to ensure that the list of entities specified on the array have been created in NROne, and also to see there exactly the expected number for each type.

### Metrics

This test is to check if the metrics specified in the spec file added to the pipeline's e2e source attribute are present on NROne. The current approach is to copy this spec file in the e2e path.

Example of metrics spec file:

```yaml
specVersion: "2"
owningTeam: integrations
integrationName: powerdns
humanReadableIntegrationName: PowerDNS
entities:
  - entityType: POWERDNS_AUTHORITATIVE
    metrics:
      - name: powerdns_authoritative_deferred_cache_actions
        type: count
        defaultResolution: 15
        unit: count
        dimensions:
          - name: type
            type: string
  - entityType: POWERDNS_Recursive
    metrics:
      - name: powerdns_recursive_deferred_cache_actions
        type: count
        defaultResolution: 15
        unit: count
        dimensions:
          - name: type
            type: string
```

This file will be parsed, getting each entity type and the metric names associated. The e2e will do a query to NROne to get all metrics with the testkey of the scenario and will fail if one is not found.

There is the possibility to skip some entity's metrics or specific metrics.

### NRQL

A list of NRQLs that will be checked in NROne, it can be any query and will fail if the result is nil or if it does not match an optional expected result.

## Support

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub.

## Contribute

We encourage your contributions to improve this action! Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.

If you have any questions, or to execute our corporate CLA (which is required if your contribution is on behalf of a company), drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

If you would like to contribute to this project, review [these guidelines](./CONTRIBUTING.md).

To all contributors, we thank you!  Without your contribution, this project would not be what it is today.  We also host a community project page dedicated to [Project Name](<LINK TO https://opensource.newrelic.com/projects/... PAGE>).

## License

newrelic-integration-e2e-action is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.
