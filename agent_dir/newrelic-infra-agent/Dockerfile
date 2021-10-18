FROM newrelic/infrastructure:latest

RUN mkdir -p /usr/local/prometheus-exporters/bin
RUN sh -c 'touch exporters/exporters_empty'

ADD integrations.d/nri-config.yml /etc/newrelic-infra/integrations.d
ADD bin/* /var/db/newrelic-infra/newrelic-integrations/bin/
ADD exporters/* /usr/local/prometheus-exporters/bin/


RUN chmod +x /var/db/newrelic-infra/newrelic-integrations/bin/* || true
RUN chmod +x /usr/local/prometheus-exporters/bin/* || true
