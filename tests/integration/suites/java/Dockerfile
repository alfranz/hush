ARG SCENARIO=pass
FROM maven:3.9-eclipse-temurin-21-alpine
ARG SCENARIO
COPY hush /usr/local/bin/hush
COPY java/pom.xml /app/pom.xml
RUN mvn -f /app/pom.xml dependency:go-offline -q
COPY java/${SCENARIO}/src/ /app/src/
WORKDIR /app
