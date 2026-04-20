package config

// TerraformProviderSource is the Terraform registry source for the upstream
// Linear provider used by Upjet code generation.
const TerraformProviderSource = "terraform-community-providers/linear"

// TerraformProviderVersion is the version of the upstream Terraform provider.
const TerraformProviderVersion = "0.5.0"

// ProviderCRDGroup is the CRD API group for all Linear resources.
const ProviderCRDGroup = "linear.crossplane.io"

// ProviderCRDAPIVersion is the CRD API version for all Linear resources.
const ProviderCRDAPIVersion = "v1alpha1"

// ProviderShortName is the short name used in resource naming.
const ProviderShortName = "linear"
