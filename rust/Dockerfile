FROM rust:1-slim-bookworm AS builder

WORKDIR /src

ENV PATH=/usr/local/cargo/bin:${PATH} \
    CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse

COPY Cargo.toml Cargo.lock ./
COPY src ./src

RUN /usr/local/cargo/bin/cargo build --release --locked

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /src/target/release/any2api-rust ./any2api-rust

ENV NEWPLATFORM2API_HOST=0.0.0.0 \
    NEWPLATFORM2API_PORT=8101 \
    NEWPLATFORM2API_ADMIN_STORE_PATH=/app/data/admin.json

EXPOSE 8101

CMD ["./any2api-rust"]