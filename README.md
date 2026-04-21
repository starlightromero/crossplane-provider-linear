# provider-linear

A [Crossplane](https://www.crossplane.io/) provider for [Linear](https://linear.app) that enables declarative management of Linear workspace resources via Kubernetes custom resources.

Generated using [Upjet](https://github.com/crossplane/upjet) from [`terraform-community-providers/linear`](https://registry.terraform.io/providers/terraform-community-providers/linear/latest/docs).

## Managed Resources

All CRDs are registered under `linear.crossplane.io/v1alpha1`.

| Kind | Terraform Resource | Description |
|---|---|---|
| `Team` | `linear_team` | Teams with workflow states, cycles, triage, estimation |
| `TeamLabel` | `linear_team_label` | Labels scoped to a team |
| `TeamWorkflow` | `linear_team_workflow` | Git automation workflow state mappings |
| `Template` | `linear_template` | Issue, project, or document templates |
| `WorkflowState` | `linear_workflow_state` | Additional workflow states for a team |
| `WorkspaceLabel` | `linear_workspace_label` | Workspace-scoped labels |
| `WorkspaceSettings` | `linear_workspace_settings` | Workspace-level settings (singleton) |
| `Workspace` | `linear_workspace` | Read-only workspace metadata (data source) |

## Installation

### Prerequisites

- Kubernetes cluster with [Crossplane](https://docs.crossplane.io/latest/software/install/) v1.14+
- A Linear API token or OAuth2 credentials

### Install the provider

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-linear
spec:
  package: ghcr.io/starlightromero/provider-linear:latest
```

```bash
kubectl apply -f provider.yaml
```

### Configure authentication

Create a Secret with your Linear API token:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: linear-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  token: lin_api_YOUR_TOKEN_HERE
```

Create a ProviderConfig referencing the Secret:

```yaml
apiVersion: linear.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: linear-credentials
      key: token
```

The provider also supports OAuth2 client credentials and OAuth2 authorization code flows. See [`examples/`](examples/) for all auth methods.

## Usage

### Create a Team

```yaml
apiVersion: linear.crossplane.io/v1alpha1
kind: Team
metadata:
  name: engineering
spec:
  forProvider:
    key: "ENG"
    name: "Engineering"
    color: "#6B5CE7"
    triage:
      enabled: true
    cycles:
      enabled: true
      startDay: 1
      duration: 2
    estimation:
      type: "linear"
  providerConfigRef:
    name: default
```

### Create a TeamLabel with a reference

```yaml
apiVersion: linear.crossplane.io/v1alpha1
kind: TeamLabel
metadata:
  name: bug-label
spec:
  forProvider:
    name: "Bug"
    color: "#eb5757"
    teamIdRef:
      name: engineering
  providerConfigRef:
    name: default
```

### Create a WorkflowState

```yaml
apiVersion: linear.crossplane.io/v1alpha1
kind: WorkflowState
metadata:
  name: in-review
spec:
  forProvider:
    name: "In Review"
    type: "started"
    position: 2.5
    color: "#f2994a"
    teamIdRef:
      name: engineering
  providerConfigRef:
    name: default
```

See [`examples/`](examples/) for all resource types.

## Authentication Methods

| Method | `spec.credentials.source` | Description |
|---|---|---|
| API Token | `Secret` | Static token from a K8s Secret |
| OAuth2 Client Credentials | `OAuth2ClientCredentials` | Server-to-server, 30-day tokens |
| OAuth2 Authorization Code | `OAuth2` | User-delegated, 24-hour tokens with auto-refresh |

Each ProviderConfig uses exactly one method. OAuth2 authorization code tokens are automatically refreshed and the new token pair is written back to the K8s Secret.

## Cross-Resource References

The provider supports Crossplane-native resource references:

| Source | Field | Target |
|---|---|---|
| TeamLabel | `teamIdRef` | Team |
| WorkflowState | `teamIdRef` | Team |
| Template | `teamIdRef` | Team |
| TeamWorkflow | `draftRef`, `startRef`, `reviewRef`, `mergeableRef`, `mergeRef` | WorkflowState |

When a referenced resource doesn't exist or isn't ready, the referencing resource sets `Synced=False` and waits.

## Field Validation

Validation runs at admission time before reaching the Linear API:

| Constraint | Fields | Rule |
|---|---|---|
| UUID | `teamId`, `parentId`, workflow state refs | Standard UUID format |
| Hex Color | `color` on Team, TeamLabel, WorkflowState, WorkspaceLabel | `#rrggbb` |
| Team Key | `key` on Team | 1-5 uppercase alphanumeric chars |
| Name Length | `name` on Team (≥2), all others (≥1) | Minimum character count |
| Enum | `autoArchivePeriod`, `autoClosePeriod`, WorkflowState `type`, Template `type` | Fixed value sets |
| JSON | Template `data` | Valid JSON |
| Range | `fiscalYearStartMonth` | Integer 0-11 |
| Immutable | TeamLabel `teamId`, WorkflowState `type` and `teamId` | Cannot change after creation |

## Reconciliation

The provider follows standard Crossplane reconciliation patterns:

- `Ready=True` / `Synced=True` after successful create/update
- Exponential backoff on API errors
- Rate limit respect (429 with `Retry-After`)
- `deletionPolicy: Orphan` leaves Linear objects intact
- `managementPolicies` for fine-grained observe/create/update/delete control
- `status.atProvider` populated with latest observed state

## Development

### Prerequisites

- Go 1.26+
- Docker (for image builds)

### Build

```bash
make build        # Build the provider binary
make build.image  # Build the container image
```

### Test

```bash
make test              # Unit + property-based tests
make test.integration  # Integration tests
make test.coverage     # Tests with coverage report
```

### Lint

```bash
make lint      # Run golangci-lint
make lint.fix  # Auto-fix lint issues
```

### Generate

```bash
make generate       # Run Upjet code generation
make generate.init  # Initialize the generation pipeline
```

### Security Scans

```bash
make scan.trivy          # Trivy filesystem scan
make validate.manifests  # kubeconform strict validation
```

## Architecture

```
┌─────────────────────────────────────────────────┐
│              Kubernetes Cluster                  │
│                                                  │
│  ┌────────────┐    ┌──────────────────────────┐  │
│  │ Crossplane │───▶│ provider-linear          │  │
│  │  Runtime   │    │  (Upjet, distroless,     │  │
│  └────────────┘    │   no-fork TF runtime)    │  │
│                    └──────────┬───────────────┘  │
│                               │                  │
│  ┌────────────┐               │                  │
│  │ K8s Secret │◀──────────────┤                  │
│  │ (API token │               │                  │
│  │  / OAuth2) │               ▼                  │
│  └────────────┘    ┌──────────────────────────┐  │
│                    │  Linear GraphQL API      │  │
│                    │  api.linear.app/graphql   │  │
│                    └──────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## CI/CD

- **CI** — Tests, Trivy (filesystem + image), CodeQL, SBOM generation on every push/PR
- **Release** — Semantic-release from conventional commits, multi-arch image to GHCR, Crossplane xpkg package, GitHub Release with SBOM
- **Renovate** — Automated dependency updates with SHA-pinned GitHub Actions

All GitHub Actions are SHA-pinned to immutable commits.

## License

[GPL-3.0](LICENSE)
