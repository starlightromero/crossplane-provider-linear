# Build stage
FROM golang:1.26.4@sha256:68cb6d68bed024785b69195b89af7ac7a444f27791435f98647edff595aa0479 AS builder

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

# Build the Linear provider plugin from our fork (includes release resources)
RUN apk add --no-cache go git && \
    git clone --depth 1 --branch feat/release-resources https://github.com/starlightromero/terraform-provider-linear.git /tmp/tf-linear && \
    cd /tmp/tf-linear && \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -o /tmp/terraform-provider-linear . && \
    mkdir -p /terraform-plugins/registry.terraform.io/terraform-community-providers/linear/0.3.7/linux_${TARGETARCH} && \
    mv /tmp/terraform-provider-linear /terraform-plugins/registry.terraform.io/terraform-community-providers/linear/0.3.7/linux_${TARGETARCH}/terraform-provider-linear_v0.3.7

# Create terraformrc for filesystem mirror
RUN printf 'provider_installation {\n  filesystem_mirror {\n    path = "/terraform-plugins"\n  }\n}\n' > /terraformrc

# Runtime stage
FROM gcr.io/distroless/static:nonroot@sha256:1c2c046bc09ed40fad370b599a0b1ae7987f55b01e247cf27a7c27cd97e5bbc7
COPY --from=builder /usr/local/bin/provider-linear /usr/local/bin/provider-linear
COPY --from=terraform /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=terraform /terraform-plugins /terraform-plugins
COPY --from=terraform /terraformrc /.terraformrc

ENV TF_CLI_CONFIG_FILE=/.terraformrc
USER 65532:65532

# Prometheus metrics endpoint
EXPOSE 8080

ENTRYPOINT ["provider-linear"]
