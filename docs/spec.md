# Crossplane Provider Linear — Spec

## Overview

A Crossplane provider for [Linear](https://linear.app) that enables
declarative management of Linear workspace resources via Kubernetes
custom resources. The provider is generated using
[Upjet](https://github.com/crossplane/upjet) from the community
Terraform provider
[terraform-community-providers/linear](https://registry.terraform.io/providers/terraform-community-providers/linear/latest/docs).

## Upstream Terraform Provider

| Field | Value |
|---|---|
| Registry | `terraform-community-providers/linear` |
| API | Linear GraphQL API (`https://api.linear.app/graphql`) |
| Auth | API token via `LINEAR_TOKEN` env var or provider config |

## Managed Resources

Each Terraform resource maps to a Crossplane Managed Resource (MR).
The CRD group is `linear.crossplane.io`.

### 1. `Team` — `linear_team`

Manages a Linear team with full lifecycle support.

| Field | Type | Required | Description |
|---|---|---|---|
| `key` | string | yes | Team key (≤5 chars, uppercase alphanumeric) |
| `name` | string | yes | Team name (≥2 chars) |
| `private` | bool | no | Team privacy (default: `false`) |
| `parentId` | string | no | Parent team UUID |
| `description` | string | no | Team description |
| `icon` | string | no | Icon (letters only) |
| `color` | string | no | Hex color (`#rrggbb`) |
| `timezone` | string | no | Timezone (default: `Etc/GMT`) |
| `enableIssueHistoryGrouping` | bool | no | default: `true` |
| `enableIssueDefaultToBottom` | bool | no | default: `false` |
| `enableThreadSummaries` | bool | no | default: `true` |
| `autoArchivePeriod` | float64 | no | Months: 1,3,6,9,12 (default: `6`) |
| `autoClosePeriod` | float64 | no | Months: 0,1,3,6,9,12 (default: `6`) |
| `autoCloseParentIssues` | bool | no | default: `false` |
| `autoCloseChildIssues` | bool | no | default: `false` |
| `triage` | object | no | `{enabled, requirePriority}` |
| `cycles` | object | no | `{enabled, startDay, duration, cooldown, upcoming, autoAddStarted, autoAddCompleted, needForActive}` |
| `estimation` | object | no | `{type, extended, allowZero, default}` |
| `backlogWorkflowState` | object | no | `{name, color, description}` (computed: `id`, `position`) |
| `unstartedWorkflowState` | object | no | Same shape as above |
| `startedWorkflowState` | object | no | Same shape as above |
| `completedWorkflowState` | object | no | Same shape as above |
| `canceledWorkflowState` | object | no | Same shape as above |

Import key: `key`

### 2. `TeamLabel` — `linear_team_label`

Manages a label scoped to a specific team.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label name (≥1 char) |
| `teamId` | string | yes | Team UUID (immutable) |
| `description` | string | no | Label description |
| `color` | string | no | Hex color |
| `parentId` | string | no | Parent label group UUID |

Import key: `label_name:team_key`

### 3. `TeamWorkflow` — `linear_team_workflow`

Manages git automation workflow states for a team.

| Field | Type | Required | Description |
|---|---|---|---|
| `key` | string | yes | Team key |
| `branch` | object | no | `{pattern, isRegex}` (computed: `id`) |
| `draft` | string | no | Workflow state UUID for draft PRs |
| `start` | string | no | Workflow state UUID for opened PRs |
| `review` | string | no | Workflow state UUID for review requests |
| `mergeable` | string | no | Workflow state UUID for mergeable PRs |
| `merge` | string | no | Workflow state UUID for merged PRs |

Import key: `team_key` or `team_key:branch_pattern:is_regex`

### 4. `Template` — `linear_template`

Manages issue, project, or document templates.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Template name (≥1 char) |
| `data` | string | yes | Template data (JSON) |
| `type` | string | no | `issue`, `project`, `document` (default: `issue`) |
| `teamId` | string | no | Team UUID (omit for workspace-level) |
| `description` | string | no | Template description |

Import key: `id`

### 5. `WorkflowState` — `linear_workflow_state`

Manages additional workflow states for a team.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | State name (≥1 char) |
| `type` | string | yes | `triage`, `backlog`, `unstarted`, `started`, `completed`, `canceled` (immutable) |
| `position` | number | yes | Sort position |
| `color` | string | yes | Hex color |
| `teamId` | string | yes | Team UUID (immutable) |
| `description` | string | no | State description |

Import key: `workflow_state_name:team_key`

### 6. `WorkspaceLabel` — `linear_workspace_label`

Manages a workspace-scoped label.

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Label name (≥1 char) |
| `description` | string | no | Label description |
| `color` | string | no | Hex color |
| `parentId` | string | no | Parent label group UUID |

Import key: `label_name`

### 7. `WorkspaceSettings` — `linear_workspace_settings`

Manages workspace-level settings (singleton resource).

| Field | Type | Required | Description |
|---|---|---|---|
| `allowMembersToInvite` | bool | no | default: `true` |
| `allowMembersToCreateTeams` | bool | no | default: `true` |
| `allowMembersToManageLabels` | bool | no | default: `true` |
| `enableGitLinkbackMessages` | bool | no | default: `true` |
| `enableGitLinkbackMessagesPublic` | bool | no | default: `false` |
| `fiscalYearStartMonth` | int | no | 0-11 (default: `0`) |
| `projects` | object | no | `{updateReminderDay, updateReminderHour, updateReminderFrequency}` |
| `initiatives` | object | no | `{enabled, updateReminderDay, updateReminderHour, updateReminderFrequency}` |
| `feed` | object | no | `{enabled, schedule}` |
| `customers` | object | no | `{enabled}` |

Import key: `id`

## Data Sources

### `Workspace` — `linear_workspace`

Read-only data source returning workspace metadata.

| Field | Type | Description |
|---|---|---|
| `id` | string | Workspace identifier |
| `name` | string | Workspace name |
| `urlKey` | string | Workspace URL key |

## Provider Configuration

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

The provider authenticates via a Linear API token stored in a
Kubernetes Secret.

## Architecture

```
┌─────────────────────────────────────────────┐
│              Kubernetes Cluster              │
│                                              │
│  ┌────────────┐    ┌──────────────────────┐  │
│  │ Crossplane │───▶│ provider-linear      │  │
│  │  Runtime   │    │  (Upjet-generated)   │  │
│  └────────────┘    └──────────┬───────────┘  │
│                               │              │
│                               ▼              │
│                    ┌──────────────────────┐  │
│                    │  Linear GraphQL API  │  │
│                    │  api.linear.app      │  │
│                    └──────────────────────┘  │
└─────────────────────────────────────────────┘
```

## Integration Target

The provider will be deployed via the `crossplane` Flux module in
Avodah's `aws-eks-modules` repository. The module will:

1. Install the provider package
2. Create a `ProviderConfig` referencing the Linear API token from
   External Secrets
3. Expose managed resources for team and workspace configuration

## Implementation Notes

- Generated via Upjet from `terraform-community-providers/linear`
- Linear API uses GraphQL — the TF provider wraps this via
  `genqlient`; Upjet handles the Crossplane ↔ Terraform bridge
- All UUID fields use the standard format
  `^[0-9a-fA-F]{8}-...-[0-9a-fA-F]{12}$`
- Color fields validate as `^#[0-9a-fA-F]{6}$`
- `WorkspaceSettings` is a singleton — only one instance per workspace
- Default workflow states (backlog, unstarted, started, completed,
  canceled) are managed inline on the `Team` resource and cannot be
  deleted
