FROM newrelic/infrastructure:latest

RUN mkdir -p /usr/local/prometheus-exporters/bin

ADD integrations.d/nri-config.yml /etc/newrelic-infra/integrations.d
ADD bin/* /var/db/newrelic-infra/newrelic-integrations/bin/
ADD exporters/* /usr/local/prometheus-exporters/bin/


RUN chmod +x /var/db/newrelic-infra/newrelic-integrations/bin/*
RUN chmod +x /usr/local/prometheus-exporters/bin/*