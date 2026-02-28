ARG SCENARIO=pass
FROM python:3.13-slim
ARG SCENARIO
RUN pip install pytest --no-cache-dir -q
COPY hush /usr/local/bin/hush
COPY python/${SCENARIO}/ /app/tests/
WORKDIR /app
