FROM golang:1.24
WORKDIR /app
COPY . .
# Compile packages and tests ahead, so the measured run prints only test output.
RUN go vet ./...
