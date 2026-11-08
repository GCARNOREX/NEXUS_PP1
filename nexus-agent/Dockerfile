FROM rust:1-bookworm AS build
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bookworm-slim
COPY --from=build /app/target/release/nexus-agent /usr/local/bin/nexus-agent
CMD ["nexus-agent"]