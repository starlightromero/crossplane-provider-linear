# 0003. Code generation pipeline and postgen patching

Date: 2026-05-29

## Status

Accepted

## Context

Upjet generates Crossplane types and controllers from a Terraform
provider schema, but the generated output requires several
modifications to work correctly with this provider:

1. The Upjet pipeline generates per-resource-group API packages (e.g.,
   `apis/team/v1alpha1/`) in addition to the consolidated
   `apis/linear/v1alpha1/` package. These produce incorrect CRD groups
   like `team.linear.crossplane.io` instead of `linear.crossplane.io`.

2. The generated webhook controller code uses a deprecated
   `ctrl.NewWebhookManagedBy(mgr).For(obj).Complete()` signature that
   doesn't compile with newer controller-runtime versions.

3. Manual resources (User) need to be wired into the generated
   `zz_setup.go` controller registration.

4. The ProviderConfig API package needs to be registered in
   `zz_register.go`.

## Decision

Implement a `cmd/generate/postgen.go` step that runs after the Upjet
pipeline and before controller-gen. It performs:

1. **Spurious group cleanup** — removes per-resource-group directories
   under `apis/`, `internal/controller/`, and `package/crds/` that
   don't belong to the `linear.crossplane.io` group.

2. **Webhook API patching** — rewrites the deprecated webhook
   registration call to the current API.

3. **Managed shim generation** — generates `GetItems()` list helpers
   and managed interface accessors for all types.

4. **Controller registration** — patches `zz_setup.go` to include the
   manual User controller.

5. **Scheme registration** — patches `zz_register.go` to include the
   ProviderConfig package.

The full pipeline (`cmd/generate/main.go`) is:
1. Upjet `PipelineRunner.Run()` — generates types + controllers
2. `runPostgen()` — applies all patches above
3. `controller-gen` — generates deepcopy methods and CRD YAML

## Consequences

- Running `go generate ./...` produces a fully working provider with
  no manual intervention needed
- The postgen step is idempotent — running it multiple times produces
  the same output
- New resources added to the Terraform provider only require a schema
  regeneration and `go generate` to appear in the Crossplane provider
- The cleanup step ensures the PR diff only contains
  `linear.crossplane.io` resources regardless of Upjet's default
  behavior
