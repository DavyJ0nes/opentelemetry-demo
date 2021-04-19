FROM openjdk:14-jdk-alpine

ARG JAR_FILE=target/*.jar
COPY ${JAR_FILE} app.jar
COPY opentelemetry-javaagent-all.jar .
EXPOSE 8080
ENTRYPOINT ["java","-javaagent:./opentelemetry-javaagent-all.jar","-jar","/app.jar"]