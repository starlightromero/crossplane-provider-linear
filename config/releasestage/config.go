// Package releasestage contains the Upjet resource configuration for linear_release_stage.
package releasestage

import ujconfig "github.com/crossplane/upjet/v2/pkg/config"

const extractorPackagePath = "github.com/crossplane/upjet/v2/pkg/resource"

// Configure adds custom resource configuration for linear_release_stage.
func Configure(p *ujconfig.Provider) {
	p.AddResourceConfigurator("linear_release_stage", func(r *ujconfig.Resource) {
		r.ShortGroup = ""
		r.Kind = "ReleaseStage"
		r.Version = "v1alpha1"

		// Cross-resource reference: pipeline_id → linear_release_pipeline.
		r.References["pipeline_id"] = ujconfig.Reference{
			TerraformName: "linear_release_pipeline",
			Extractor:     extractorPackagePath + ".ExtractParamPath(\"id\", true)",
		}
	})
}
