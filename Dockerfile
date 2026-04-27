# Build stage
FROM golang:1.26@sha256:b54cbf583d390341599d7bcbc062425c081105cc5ef6d170ced98ef9d047c716 AS builder

ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /workspace

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -ldflags "-s -w -X github.com/avodah-inc/provider-linear/internal/version.Version=${VERSION}" \
    -o /usr/local/bin/provider-linear ./cmd/provider/

# Runtime stage — distroless, no shell
FROM gcr.io/distroless/static:nonroot@sha256:e3f945647ffb95b5839c07038d64f9811adf17308b9121d8a2b87b6a22a80a39

COPY --from=builder /usr/local/bin/provider-linear /usr/local/bin/provider-linear

USER 65532:65532

# Prometheus metrics endpoint
EXPOSE 8080

ENTRYPOINT ["provider-linear"]
