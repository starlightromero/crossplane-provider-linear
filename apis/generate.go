// Package apis contains the API types for provider-linear.
//
// Generate CRD manifests from API types into package/crds/.
//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen crd:allowDangerousTypes=true,crdVersions=v1 paths=./... output:artifacts:config=../package/crds
package apis
