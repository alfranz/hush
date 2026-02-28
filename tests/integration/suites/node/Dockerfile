ARG SCENARIO=pass
FROM node:22-slim
ARG SCENARIO
COPY hush /usr/local/bin/hush
COPY node/${SCENARIO}/ /app/tests/
WORKDIR /app/tests
