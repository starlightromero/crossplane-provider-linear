# 0001. Use namespaced scope for all managed resources

Date: 2026-05-29

## Status

Accepted

## Context

Crossplane providers historically default to cluster-scoped CRDs for
managed resources. This was the Upjet default and how the provider was
initially generated. However, cluster-scoped resources present
challenges for multi-tenant clusters:

- No RBAC isolation between teams — any user with access to the CRD
  can read/modify all instances across the cluster
- No organizational boundary — resources from different teams/projects
  are mixed in a flat namespace
- Harder to implement least-privilege access patterns

## Decision

Switch all managed resources to `Namespaced` scope. Only
`ProviderConfig` and `ProviderConfigUsage` remain cluster-scoped, as
required by the Crossplane runtime.

Implementation:
- Set `Scope: tjtypes.CRDScopeNamespaced` on the `PipelineRunner` in
  `cmd/generate/main.go`
- Migrate manual resources (`User`, `TeamMembership`) to
  `v2.ManagedResourceSpec` which uses `ProviderConfigReference` and
  `LocalSecretReference` (namespaced types)
- Update postgen managed shims to use namespaced interface types

## Consequences

- Teams can scope Linear resources to their own namespace with
  standard Kubernetes RBAC
- Existing deployments using cluster-scoped CRDs will need to migrate
  (this is a breaking change for the CRD schema)
- Connection secrets are written to the same namespace as the managed
  resource (LocalSecretReference) rather than any arbitrary namespace
- ProviderConfig remains cluster-scoped, so a single provider
  configuration can serve multiple namespaces
