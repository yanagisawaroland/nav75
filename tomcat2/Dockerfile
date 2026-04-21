FROM harbor.devops.local:7581/devops/tomcat:9.0-jdk11

# JMX 监控
COPY jmx_prometheus_javaagent.jar /opt/
COPY jmx-config.yaml /opt/

# ROOT.war = 部署在根路径 /，无 context path
# 访问 /health 而不是 /tomcat2/health
# 彻底解决 canary Ingress rewrite 404 问题
COPY target/tomcat2.war /usr/local/tomcat/webapps/ROOT.war

ENV CATALINA_OPTS="-javaagent:/opt/jmx_prometheus_javaagent.jar=9404:/opt/jmx-config.yaml \
    -Xms64m -Xmx128m"

EXPOSE 8080 9404
