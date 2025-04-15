ARG GO_VERSION=1

FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

#######################################
FROM debian:bookworm-slim

HEALTHCHECK --interval=1m --timeout=10s --retries=3 \
  CMD curl --fail http://localhost:8080/health || exit 1

WORKDIR /app

# Install deps for certificates
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    xz-utils=5.4.1-1 \
    curl=7.88.1-10+deb12u12 \
    ca-certificates=20230311 \
    fonts-inter=4.0~beta7+ds-1 && \
  rm -rf /var/lib/apt/lists/*

# Install typst from the binary
RUN curl -L https://github.com/typst/typst/releases/latest/download/typst-x86_64-unknown-linux-musl.tar.xz -o typst.tar.xz \
  && tar -xf typst.tar.xz \
  && mv typst-x86_64-unknown-linux-musl/typst /usr/local/bin/ \
  && rm -rf typst-x86_64-unknown-linux-musl typst.tar.xz

COPY --from=builder /run-app /usr/local/bin/
COPY assets assets

# Create a non-root user and group
RUN groupadd -r appgroup && useradd -r -g appgroup -G appgroup -m -d /app -s /sbin/nologin appuser

# Switch to the non-root user
USER appuser

CMD ["run-app"]
