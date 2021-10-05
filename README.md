[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# newrelic-integration-e2e-action

End to end testing for newrelic integrations to ensure the integrations are correctly executed by the agent, and the service metrics and entities are correctly sent to NROne.

New Relic has two kinds of integrations:
- Custom made integrations (e2e testing still not supported)
- Integrations based on prometheus exporters

## Steps the of the e2e action

- It reads the e2e test descriptor file/s that must be passed as an argument to the action.
- For each scenario present on the descriptor:  
    - It installs the infrastructure agent & the required packages.
    - Launch services dependencies (e.g. a docker-compose ) if they’re required.
    - It verifies that required services are up & running
    - It creates a config file with the details in the descriptor. 
    - Adds a custom-attribute to the config:
        - Composed by the current commit sha + a new 10 alphanumeric-random digit on each scenario.
        - The tests will look for this label to fetch the metrics and the entities from the New Relic backend.
    - The runner executes the tests one by one, checking that metrics &/or entities are being created correctly. and if fails retries after n seconds.
    - If the test fails, it's retried after the `retry_seconds` (default 30s) and up to the `retry_attempts` (default 10) defined for the action. 
    - It stops & removes the services (if they are required).
    - If `verbose` is true it logs the agent logs with other debug information.
- The action is completed.


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
          go-version: 1.17
      - name: build-powerdns
        run: make build-powerdns
      - name: e2e-test
        uses: newrelic/newrelic-integration-e2e-action@v1
        with:
          spec_path: exporters/powerdns/e2e/e2e_spec.yml
          account_id: ${{ secrets.ACCOUNT_ID }}
          api_key: ${{ secrets.API_KEY }}
          license_key: ${{ secrets.LICENSE_KEY }} 
```

## Spec file for the e2e 
Example:
```yaml
description: |
  End-to-end tests for PowerDNS integration

agent:
  integrations:
    nri-prometheus:  bin/nri-prometheus

scenarios:
  - description: |
      This scenario will verify that metrics froms PDNS authoritative
      are correcly collected.
    before:
      - docker-compose -f "deps/docker-compose.yml" up -d
    after:
      - docker-compose -f "deps/docker-compose.yml" up -d
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
          # additionals: ""
```

## Types of test
- ENTITIES
- METRICS
- NRQL

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