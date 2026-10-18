FROM python:3.12-slim
RUN pip install --no-cache-dir pytest==8.3.4
WORKDIR /app
COPY . .
# Warm up imports so the measured run prints only test output.
RUN python -c "import pytest" && python -m compileall -q shop tests
