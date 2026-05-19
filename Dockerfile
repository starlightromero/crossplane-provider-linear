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

# Download Terraform binary and Linear provider plugin
FROM alpine:3.21 AS terraform
ARG TARGETARCH
RUN apk add --no-cache curl unzip && \
    curl -fsSL "https://releases.hashicorp.com/terraform/1.12.0/terraform_1.12.0_linux_${TARGETARCH}.zip" -o /tmp/terraform.zip && \
    unzip /tmp/terraform.zip -d /usr/local/bin/ && \
    chmod +x /usr/local/bin/terraform

RUN mkdir -p /terraform-plugins/registry.terraform.io/terraform-community-providers/linear/0.3.7/linux_${TARGETARCH} && \
    curl -fsSL "https://github.com/terraform-community-providers/terraform-provider-linear/releases/download/v0.3.7/terraform-provider-linear_0.3.7_linux_${TARGETARCH}.zip" -o /tmp/provider.zip && \
    unzip /tmp/provider.zip -d /terraform-plugins/registry.terraform.io/terraform-community-providers/linear/0.3.7/linux_${TARGETARCH}/

# Create terraformrc for filesystem mirror
RUN printf 'provider_installation {\n  filesystem_mirror {\n    path = "/terraform-plugins"\n  }\n}\n' > /terraformrc

# Runtime stage
FROM gcr.io/distroless/static:nonroot@sha256:e3f945647ffb95b5839c07038d64f9811adf17308b9121d8a2b87b6a22a80a39
COPY --from=builder /usr/local/bin/provider-linear /usr/local/bin/provider-linear
COPY --from=terraform /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=terraform /terraform-plugins /terraform-plugins
COPY --from=terraform /terraformrc /.terraformrc

ENV TF_CLI_CONFIG_FILE=/.terraformrc
USER 65532:65532

# Prometheus metrics endpoint
EXPOSE 8080

ENTRYPOINT ["provider-linear"]
