FROM openjdk:14-jdk-alpine

ARG otel_agent_version=v1.1.0
ARG JAR_FILE=target/otel-0.0.1-SNAPSHOT.jar

ADD "https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/$otel_agent_version/opentelemetry-javaagent-all.jar" /opt/opentelemetry-agent.jar
COPY ${JAR_FILE} app.jar

EXPOSE 8080
ENTRYPOINT ["java","-javaagent:/opt/opentelemetry-agent.jar","-jar","/app.jar"]