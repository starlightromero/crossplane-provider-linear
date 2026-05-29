// Package releasepipeline contains the Upjet resource configuration for linear_release_pipeline.
package releasepipeline

import ujconfig "github.com/crossplane/upjet/v2/pkg/config"

// Configure adds custom resource configuration for linear_release_pipeline.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_release_pipeline", func(r *ujconfig.Resource) {
		r.ShortGroup = ""
		r.Kind = "ReleasePipeline"
		r.Version = "v1alpha1"
	})
}
