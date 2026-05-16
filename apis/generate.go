// Package apis contains the API types for provider-linear.
//
// Generate CRD manifests from API types into package/crds/.
//go:generate go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=../hack/boilerplate.go.txt paths=./... crd:allowDangerousTypes=true,crdVersions=v1 output:artifacts:config=../package/crds
package apis
