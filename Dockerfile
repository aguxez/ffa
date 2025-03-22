ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .

#######################################
FROM debian:bookworm-slim

# Install deps for certificates
RUN apt-get update && \
  apt-get install -y --no-install-recommends xz-utils curl ca-certificates fonts-inter && \
  rm -rf /var/lib/apt/lists/*

# Install typst from the binary
RUN curl -L https://github.com/typst/typst/releases/latest/download/typst-x86_64-unknown-linux-musl.tar.xz -o typst.tar.xz \
  && tar -xf typst.tar.xz \
  && mv typst-x86_64-unknown-linux-musl/typst /usr/local/bin/ \
  && rm -rf typst-x86_64-unknown-linux-musl typst.tar.xz

COPY --from=builder /run-app /usr/local/bin/
COPY assets assets

CMD ["run-app"]
