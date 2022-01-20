FROM alpine:latest as builder
ARG NRJMX_VERSION
WORKDIR /
ADD http://download.newrelic.com/infrastructure_agent/binaries/linux/noarch/nrjmx_linux_${NRJMX_VERSION}_noarch.tar.gz ./
RUN tar -xzf nrjmx_linux_${NRJMX_VERSION}_noarch.tar.gz

FROM newrelic/infrastructure:latest

RUN apk --update add openjdk8-jre
COPY --from=builder /usr/bin/nrjmx /usr/bin/nrjmx
COPY --from=builder /usr/bin/nrjmx.jar /usr/bin/nrjmx.jar
COPY --from=builder /usr/bin/jmxterm.jar /usr/bin/jmxterm.jar
