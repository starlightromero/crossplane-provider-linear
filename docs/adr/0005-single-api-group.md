# 0005. Single API group for all resources

Date: 2026-05-29

## Status

Accepted

## Context

Upjet's default behavior generates a separate API group per Terraform
resource (e.g., `team.linear.crossplane.io`,
`template.linear.crossplane.io`). This produces many CRD groups for a
provider that manages a single external service.

Other Crossplane providers (AWS, GCP, Azure) use per-service groups
because they manage hundreds of resources across distinct services.
The Linear provider manages ~10 resources in a single API with a
unified authentication model.

## Decision

Use a single API group `linear.crossplane.io` for all resources by
setting `r.ShortGroup = ""` on every resource configurator. This
collapses all resources into one group with API version `v1alpha1`.

The postgen cleanup step removes any spurious per-resource-group
directories that Upjet generates alongside the consolidated package.

## Consequences

- All Linear resources share one API group, making discovery simple:
  `kubectl api-resources --api-group=linear.crossplane.io`
- RBAC rules can grant access to all Linear resources with a single
  group wildcard
- If the provider grows significantly (unlikely for Linear), resources
  could be split into subgroups later — but this would be a breaking
  API change
- The `ShortGroup = ""` convention must be set on every new resource
  configurator added to the provider
