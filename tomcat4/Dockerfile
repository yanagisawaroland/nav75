FROM harbor.devops.local:7581/devops/tomcat:9.0-jdk11

COPY jmx_prometheus_javaagent.jar /opt/
COPY jmx-config.yaml /opt/
COPY opentelemetry-javaagent.jar /opt/

COPY target/tomcat4.war /usr/local/tomcat/webapps/ROOT.war

ENV CATALINA_OPTS="-javaagent:/opt/jmx_prometheus_javaagent.jar=9404:/opt/jmx-config.yaml \
    -javaagent:/opt/opentelemetry-javaagent.jar \
    -Xms64m -Xmx128m"

EXPOSE 8080 9404
