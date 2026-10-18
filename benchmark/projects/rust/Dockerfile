FROM rust:1.89-slim
WORKDIR /app
COPY . .
# Compile tests ahead, so the measured run prints only test output.
RUN cargo test --no-run
