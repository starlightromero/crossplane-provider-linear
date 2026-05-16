# Build stage
FROM golang:1.26.3@sha256:313faae491b410a35402c05d35e7518ae99103d957308e940e1ae2cfa0aac29b AS builder

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

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
