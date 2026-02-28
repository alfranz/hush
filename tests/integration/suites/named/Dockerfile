FROM python:3.13-slim
RUN pip install pytest --no-cache-dir -q
COPY hush /usr/local/bin/hush
COPY named/pass/ /app/pass/
COPY named/fail/ /app/fail/
COPY named/hush.yaml /app/.hush.yaml
WORKDIR /app
