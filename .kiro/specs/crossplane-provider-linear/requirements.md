# Requirements Document

## Introduction

A Crossplane provider for Linear that enables declarative management of Linear workspace resources via Kubernetes custom resources. The provider is generated using Upjet from the community Terraform provider `terraform-community-providers/linear`. It exposes seven managed resources and one data source under the CRD group `linear.crossplane.io`, authenticating via either a Linear API token or OAuth2 credentials stored in Kubernetes Secrets.

## Glossary

- **Provider**: The Crossplane provider-linear binary that reconciles Linear custom resources against the Linear GraphQL API
- **ProviderConfig**: A Crossplane resource that holds authentication credentials (Linear API token or OAuth2 credentials) for the Provider
- **Managed_Resource**: A Kubernetes custom resource reconciled by the Provider to create, update, or delete a corresponding Linear API object
- **Data_Source**: A read-only Crossplane resource that fetches metadata from the Linear API without creating or modifying objects
- **Upjet**: The Crossplane code generator that bridges Terraform providers into Crossplane providers
- **Linear_API**: The Linear GraphQL API at `https://api.linear.app/graphql`
- **CRD**: Kubernetes Custom Resource Definition that extends the Kubernetes API with Linear resource types
- **Team**: A Linear team resource managing workflow states, cycles, estimation, and triage settings
- **TeamLabel**: A label scoped to a specific Linear team
- **TeamWorkflow**: Git automation workflow state mappings for a team
- **Template**: An issue, project, or document template in Linear
- **WorkflowState**: An additional workflow state for a team beyond the defaults
- **WorkspaceLabel**: A label scoped to the entire Linear workspace
- **WorkspaceSettings**: Singleton workspace-level configuration settings
- **Workspace**: Read-only workspace metadata data source
- **UUID**: A universally unique identifier in the format `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
- **Hex_Color**: A color string in the format `#rrggbb` matching `^#[0-9a-fA-F]{6}$`
- **Reconciliation**: The Crossplane control loop that observes desired state from a Managed_Resource and drives the Linear_API to match
- **OAuth2_Authorization_Code**: The Linear OAuth2 flow where a user authorizes the application, receives an authorization code, and exchanges it for short-lived access and refresh tokens via `https://api.linear.app/oauth/token`
- **OAuth2_Client_Credentials**: The Linear OAuth2 flow for server-to-server communication where the application exchanges its `client_id` and `client_secret` for an app actor access token valid for 30 days, without user interaction
- **Access_Token**: A short-lived bearer token (24-hour expiry for authorization code flow, 30-day expiry for client credentials flow) used to authenticate Linear_API requests
- **Refresh_Token**: A token issued alongside an Access_Token in the authorization code flow, used to obtain a new Access_Token when the current one expires

## Requirements

### Requirement 1: Provider Authentication

**User Story:** As a platform engineer, I want the Provider to authenticate with the Linear_API using either a static API token or OAuth2 credentials stored in Kubernetes Secrets, so that I can choose the authentication method that best fits my security posture and integration needs.

#### Acceptance Criteria — API Token Authentication

1. WHEN a ProviderConfig specifies `spec.credentials.source: Secret` with a secret containing a Linear API token, THE Provider SHALL authenticate all Linear_API requests using that token as a Bearer token
2. IF the referenced Kubernetes Secret does not exist, THEN THE Provider SHALL set the ProviderConfig status condition to `Ready=False` with a reason indicating the missing secret
3. IF the Linear API token is invalid or expired, THEN THE Provider SHALL set the ProviderConfig status condition to `Ready=False` with a reason indicating authentication failure
4. THE Provider SHALL read the API token from the key specified in `spec.credentials.secretRef.key` of the ProviderConfig

#### Acceptance Criteria — OAuth2 Client Credentials Authentication

1. WHEN a ProviderConfig specifies `spec.credentials.source: OAuth2ClientCredentials`, THE Provider SHALL authenticate using the OAuth2 client credentials flow by exchanging `client_id` and `client_secret` for an Access_Token via `POST https://api.linear.app/oauth/token` with `grant_type=client_credentials`
2. THE Provider SHALL read the `client_id` and `client_secret` from the Kubernetes Secret referenced in `spec.credentials.secretRef`, using the keys `clientId` and `clientSecret` respectively
3. THE Provider SHALL include the required `scope` parameter (comma-separated list of scopes) when requesting a client credentials token, defaulting to `read,write` if not specified in the ProviderConfig
4. WHEN the client credentials Access_Token expires or the Linear_API returns a 401 response, THE Provider SHALL automatically request a new Access_Token using the client credentials flow
5. THE Provider SHALL cache the client credentials Access_Token in memory and reuse it for subsequent Linear_API requests until it expires or is rejected

#### Acceptance Criteria — OAuth2 Authorization Code Authentication

1. WHEN a ProviderConfig specifies `spec.credentials.source: OAuth2`, THE Provider SHALL authenticate using a pre-obtained OAuth2 Access_Token and Refresh_Token stored in a Kubernetes Secret
2. THE Provider SHALL read the `access_token` and `refresh_token` from the Kubernetes Secret referenced in `spec.credentials.secretRef`, along with `client_id` and `client_secret` required for token refresh
3. WHEN the OAuth2 Access_Token expires (24-hour lifetime), THE Provider SHALL automatically refresh it by calling `POST https://api.linear.app/oauth/token` with `grant_type=refresh_token`, the current Refresh_Token, `client_id`, and `client_secret`
4. AFTER a successful token refresh, THE Provider SHALL update the Kubernetes Secret with the new Access_Token and Refresh_Token returned by the Linear_API
5. IF a token refresh fails, THEN THE Provider SHALL set the ProviderConfig status condition to `Ready=False` with a reason indicating the refresh failure and the need for re-authorization

#### Acceptance Criteria — General

1. THE Provider SHALL support exactly one authentication method per ProviderConfig, determined by the `spec.credentials.source` field
2. IF the `spec.credentials.source` field specifies an unsupported value, THEN THE Provider SHALL reject the ProviderConfig with a validation error

### Requirement 2: Provider Packaging and Installation

**User Story:** As a platform engineer, I want the Provider to be installable as a standard Crossplane provider package, so that I can deploy it using existing Crossplane tooling and Flux modules.

#### Acceptance Criteria

1. THE Provider SHALL be distributed as an OCI-compliant Crossplane provider package
2. WHEN the Provider package is installed, THE Provider SHALL register CRDs for all seven Managed_Resources and the Workspace Data_Source under the group `linear.crossplane.io`
3. THE Provider SHALL run as a distroless container image with no shell access
4. WHEN the Provider starts, THE Provider SHALL become ready within 60 seconds on a healthy cluster
5. THE Provider SHALL expose Prometheus metrics for reconciliation latency and error counts

### Requirement 3: Team Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage Linear teams via Kubernetes custom resources, so that team configuration is version-controlled and reproducible.

#### Acceptance Criteria

1. WHEN a Team resource is created with a valid `key` and `name`, THE Provider SHALL create a corresponding team in the Linear_API
2. WHEN a Team resource spec is updated, THE Provider SHALL update the corresponding Linear team to match the desired state
3. WHEN a Team resource is deleted, THE Provider SHALL delete the corresponding team from the Linear_API
4. THE Provider SHALL validate that the `key` field contains at most 5 uppercase alphanumeric characters
5. THE Provider SHALL validate that the `name` field contains at least 2 characters
6. WHEN a Team resource specifies a `color` field, THE Provider SHALL validate that the value matches the Hex_Color format
7. WHEN a Team resource specifies an `autoArchivePeriod`, THE Provider SHALL validate that the value is one of 1, 3, 6, 9, or 12
8. WHEN a Team resource specifies an `autoClosePeriod`, THE Provider SHALL validate that the value is one of 0, 1, 3, 6, 9, or 12
9. WHEN a Team resource specifies inline workflow states (backlog, unstarted, started, completed, canceled), THE Provider SHALL manage those states as part of the Team lifecycle
10. WHEN a Team resource specifies `triage` settings, THE Provider SHALL configure triage with the `enabled` and `requirePriority` sub-fields
11. WHEN a Team resource specifies `cycles` settings, THE Provider SHALL configure cycle automation with all sub-fields (enabled, startDay, duration, cooldown, upcoming, autoAddStarted, autoAddCompleted, needForActive)
12. WHEN a Team resource specifies `estimation` settings, THE Provider SHALL configure estimation with the `type`, `extended`, `allowZero`, and `default` sub-fields
13. THE Provider SHALL populate the Team resource status with the Linear-assigned team `id` after successful creation
14. IF the Linear_API returns an error during Team reconciliation, THEN THE Provider SHALL set the Team status condition to `Synced=False` with the error message

### Requirement 4: TeamLabel Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage team-scoped labels, so that label taxonomies are consistent and reproducible across environments.

#### Acceptance Criteria

1. WHEN a TeamLabel resource is created with a valid `name` and `teamId`, THE Provider SHALL create a corresponding label scoped to that team in the Linear_API
2. WHEN a TeamLabel resource spec is updated, THE Provider SHALL update the corresponding label in the Linear_API
3. WHEN a TeamLabel resource is deleted, THE Provider SHALL delete the corresponding label from the Linear_API
4. THE Provider SHALL validate that the `name` field contains at least 1 character
5. THE Provider SHALL treat the `teamId` field as immutable after creation
6. WHEN a TeamLabel resource specifies a `color` field, THE Provider SHALL validate that the value matches the Hex_Color format
7. WHEN a TeamLabel resource specifies a `parentId`, THE Provider SHALL associate the label with the specified parent label group
8. IF the referenced `teamId` does not exist in the Linear_API, THEN THE Provider SHALL set the TeamLabel status condition to `Synced=False` with a descriptive error

### Requirement 5: TeamWorkflow Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage git automation workflow mappings for teams, so that PR-to-issue state transitions are codified and version-controlled.

#### Acceptance Criteria

1. WHEN a TeamWorkflow resource is created with a valid `key`, THE Provider SHALL create or update the git automation workflow configuration for that team in the Linear_API
2. WHEN a TeamWorkflow resource spec is updated, THE Provider SHALL update the corresponding workflow configuration in the Linear_API
3. WHEN a TeamWorkflow resource is deleted, THE Provider SHALL remove the git automation workflow configuration from the Linear_API
4. WHEN a TeamWorkflow resource specifies a `branch` object, THE Provider SHALL configure the branch pattern with `pattern` and `isRegex` sub-fields
5. WHEN a TeamWorkflow resource specifies workflow state UUIDs for `draft`, `start`, `review`, `mergeable`, or `merge`, THE Provider SHALL validate that each value matches the UUID format
6. IF a referenced workflow state UUID does not exist for the team, THEN THE Provider SHALL set the TeamWorkflow status condition to `Synced=False` with a descriptive error

### Requirement 6: Template Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage Linear templates, so that issue, project, and document templates are standardized across the workspace.

#### Acceptance Criteria

1. WHEN a Template resource is created with a valid `name` and `data`, THE Provider SHALL create a corresponding template in the Linear_API
2. WHEN a Template resource spec is updated, THE Provider SHALL update the corresponding template in the Linear_API
3. WHEN a Template resource is deleted, THE Provider SHALL delete the corresponding template from the Linear_API
4. THE Provider SHALL validate that the `name` field contains at least 1 character
5. THE Provider SHALL validate that the `data` field contains valid JSON
6. WHEN a Template resource specifies a `type` field, THE Provider SHALL validate that the value is one of `issue`, `project`, or `document`
7. WHEN a Template resource omits the `type` field, THE Provider SHALL default the type to `issue`
8. WHEN a Template resource specifies a `teamId`, THE Provider SHALL scope the template to that team
9. WHEN a Template resource omits the `teamId`, THE Provider SHALL create a workspace-level template

### Requirement 7: WorkflowState Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage additional workflow states for teams, so that custom issue lifecycle stages are version-controlled.

#### Acceptance Criteria

1. WHEN a WorkflowState resource is created with valid `name`, `type`, `position`, `color`, and `teamId`, THE Provider SHALL create a corresponding workflow state in the Linear_API
2. WHEN a WorkflowState resource spec is updated, THE Provider SHALL update the corresponding workflow state in the Linear_API
3. WHEN a WorkflowState resource is deleted, THE Provider SHALL delete the corresponding workflow state from the Linear_API
4. THE Provider SHALL validate that the `name` field contains at least 1 character
5. THE Provider SHALL validate that the `type` field is one of `triage`, `backlog`, `unstarted`, `started`, `completed`, or `canceled`
6. THE Provider SHALL treat the `type` field as immutable after creation
7. THE Provider SHALL treat the `teamId` field as immutable after creation
8. THE Provider SHALL validate that the `color` field matches the Hex_Color format
9. IF the referenced `teamId` does not exist in the Linear_API, THEN THE Provider SHALL set the WorkflowState status condition to `Synced=False` with a descriptive error

### Requirement 8: WorkspaceLabel Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage workspace-scoped labels, so that organization-wide label standards are codified.

#### Acceptance Criteria

1. WHEN a WorkspaceLabel resource is created with a valid `name`, THE Provider SHALL create a corresponding workspace-scoped label in the Linear_API
2. WHEN a WorkspaceLabel resource spec is updated, THE Provider SHALL update the corresponding label in the Linear_API
3. WHEN a WorkspaceLabel resource is deleted, THE Provider SHALL delete the corresponding label from the Linear_API
4. THE Provider SHALL validate that the `name` field contains at least 1 character
5. WHEN a WorkspaceLabel resource specifies a `color` field, THE Provider SHALL validate that the value matches the Hex_Color format
6. WHEN a WorkspaceLabel resource specifies a `parentId`, THE Provider SHALL associate the label with the specified parent label group

### Requirement 9: WorkspaceSettings Managed Resource

**User Story:** As a platform engineer, I want to declaratively manage workspace-level settings, so that workspace configuration is reproducible and auditable.

#### Acceptance Criteria

1. WHEN a WorkspaceSettings resource is created, THE Provider SHALL apply the specified settings to the Linear workspace via the Linear_API
2. WHEN a WorkspaceSettings resource spec is updated, THE Provider SHALL update the workspace settings in the Linear_API to match the desired state
3. THE Provider SHALL enforce that only one WorkspaceSettings resource exists per workspace (singleton pattern)
4. IF a second WorkspaceSettings resource is created for the same workspace, THEN THE Provider SHALL set the status condition to `Synced=False` with a reason indicating a singleton violation
5. WHEN a WorkspaceSettings resource specifies `fiscalYearStartMonth`, THE Provider SHALL validate that the value is an integer between 0 and 11 inclusive
6. WHEN a WorkspaceSettings resource specifies `projects` settings, THE Provider SHALL configure project update reminders with `updateReminderDay`, `updateReminderHour`, and `updateReminderFrequency` sub-fields
7. WHEN a WorkspaceSettings resource specifies `initiatives` settings, THE Provider SHALL configure initiative settings with `enabled` and update reminder sub-fields
8. WHEN a WorkspaceSettings resource specifies `feed` settings, THE Provider SHALL configure feed with `enabled` and `schedule` sub-fields
9. WHEN a WorkspaceSettings resource specifies `customers` settings, THE Provider SHALL configure customer features with the `enabled` sub-field
10. WHEN a WorkspaceSettings resource is deleted, THE Provider SHALL reset workspace settings to Linear defaults

### Requirement 10: Workspace Data Source

**User Story:** As a platform engineer, I want to read workspace metadata via a Kubernetes custom resource, so that I can reference workspace identity in other resource configurations.

#### Acceptance Criteria

1. WHEN a Workspace data source resource is created, THE Provider SHALL fetch the workspace metadata from the Linear_API
2. THE Provider SHALL populate the Workspace status with the `id`, `name`, and `urlKey` fields from the Linear_API response
3. THE Provider SHALL periodically refresh the Workspace data to reflect upstream changes
4. THE Provider SHALL treat the Workspace resource as read-only and reject spec updates that attempt to modify Linear workspace properties

### Requirement 11: Upjet Code Generation

**User Story:** As a provider developer, I want the Provider codebase to be generated via Upjet from the upstream Terraform provider, so that resource schemas and reconciliation logic stay aligned with the Terraform provider.

#### Acceptance Criteria

1. THE Provider SHALL use Upjet to generate Crossplane controllers and CRDs from the `terraform-community-providers/linear` Terraform provider
2. THE Provider SHALL include an Upjet configuration that maps each of the seven Terraform resources to a Crossplane Managed_Resource
3. THE Provider SHALL include an Upjet configuration that maps the `linear_workspace` Terraform data source to a Crossplane Data_Source
4. WHEN the upstream Terraform provider is updated, THE Provider SHALL support regeneration via `make generate` to incorporate upstream changes
5. THE Provider SHALL configure external name annotations for each resource using the import keys defined in the spec (e.g., `key` for Team, `label_name:team_key` for TeamLabel)

### Requirement 12: Reconciliation Behavior

**User Story:** As a platform engineer, I want the Provider to follow standard Crossplane reconciliation patterns, so that resource lifecycle management is predictable and observable.

#### Acceptance Criteria

1. THE Provider SHALL set the `Ready` status condition to `True` on each Managed_Resource after successful creation or update in the Linear_API
2. THE Provider SHALL set the `Synced` status condition to `True` when the observed state matches the desired state
3. IF the Linear_API is unreachable, THEN THE Provider SHALL set the `Synced` status condition to `False` and retry with exponential backoff
4. IF the Linear_API returns a rate limit response, THEN THE Provider SHALL respect the rate limit headers and retry after the indicated delay
5. WHEN a Managed_Resource specifies a `deletionPolicy` of `Orphan`, THE Provider SHALL remove the Kubernetes resource without deleting the corresponding Linear object
6. THE Provider SHALL support the `managementPolicies` field for fine-grained control over observe, create, update, and delete operations
7. THE Provider SHALL populate the `status.atProvider` section of each Managed_Resource with the latest observed state from the Linear_API

### Requirement 13: Field Validation

**User Story:** As a platform engineer, I want the Provider to validate resource fields at admission time, so that invalid configurations are rejected before reaching the Linear_API.

#### Acceptance Criteria

1. THE Provider SHALL validate all UUID fields against the UUID format pattern
2. THE Provider SHALL validate all color fields against the Hex_Color format pattern
3. THE Provider SHALL enforce minimum length constraints on name fields as specified per resource (2 chars for Team name, 1 char for all other names)
4. THE Provider SHALL enforce maximum length constraints on the Team `key` field (5 characters)
5. THE Provider SHALL enforce enum constraints on fields with fixed value sets (WorkflowState type, Template type, autoArchivePeriod, autoClosePeriod)
6. THE Provider SHALL enforce immutability on fields marked as immutable (TeamLabel teamId, WorkflowState type, WorkflowState teamId)
7. WHEN a validation error occurs, THE Provider SHALL return a descriptive Kubernetes admission error identifying the invalid field and the constraint violated

### Requirement 14: Cross-Resource References

**User Story:** As a platform engineer, I want to reference other Crossplane-managed Linear resources by selector or reference, so that I can compose resources without hardcoding UUIDs.

#### Acceptance Criteria

1. WHEN a Managed_Resource field accepts a UUID reference to another Linear resource, THE Provider SHALL support Crossplane resource references via `ref` and `selector` fields
2. WHEN a TeamLabel resource references a Team via `teamIdRef`, THE Provider SHALL resolve the reference to the Team external ID
3. WHEN a WorkflowState resource references a Team via `teamIdRef`, THE Provider SHALL resolve the reference to the Team external ID
4. WHEN a TeamWorkflow resource references WorkflowState resources for `draft`, `start`, `review`, `mergeable`, or `merge`, THE Provider SHALL support references via `ref` and `selector` fields
5. IF a referenced resource does not exist or is not ready, THEN THE Provider SHALL set the referencing resource status condition to `Synced=False` and wait for the reference to resolve

### Requirement 15: Flux Integration Deployment

**User Story:** As a platform engineer, I want the Provider to be deployable via the crossplane Flux module in the aws-eks-modules repository, so that it integrates with the existing GitOps workflow.

#### Acceptance Criteria

1. THE Provider SHALL be installable via a Crossplane `Provider` manifest referencing the OCI package
2. THE Provider SHALL support ProviderConfig creation with credentials sourced from a Kubernetes Secret populated by External Secrets Operator
3. WHEN deployed via the Flux module, THE Provider SHALL reconcile all Managed_Resources within the `crossplane-system` namespace by default
4. THE Provider SHALL include example manifests for each of the seven Managed_Resources and the Workspace Data_Source

### Requirement 16: Security and Compliance

**User Story:** As a platform engineer, I want the Provider to meet security and compliance standards, so that it passes Trivy and SonarQube scans without critical or high findings.

#### Acceptance Criteria

1. THE Provider SHALL produce zero critical or high severity findings in Trivy filesystem scans
2. THE Provider SHALL produce zero critical or high severity findings in SonarQube analysis
3. THE Provider SHALL run as a non-root user in the container
4. THE Provider SHALL use a distroless base image with no shell access
5. THE Provider SHALL not embed secrets or credentials in the container image or CRD definitions
6. THE Provider SHALL validate Kubernetes resource manifests with kubeconform using strict mode and CRD schema locations
