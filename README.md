[![Community Project header](https://github.com/newrelic/opensource-website/raw/master/src/images/categories/Community_Project.png)](https://opensource.newrelic.com/oss-category/#community-project)

# newrelic-integration-e2e-action

End to end testing for newrelic integrations. 

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