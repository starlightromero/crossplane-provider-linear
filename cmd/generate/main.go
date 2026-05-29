// Package main orchestrates the full code generation pipeline:
// 1. Run Upjet pipeline (types + controllers)
// 2. Apply post-generation patches (webhook API compat, managed shims)
// 3. Run controller-gen (deepcopy + CRDs)
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/crossplane/upjet/v2/pkg/pipeline"
	tjtypes "github.com/crossplane/upjet/v2/pkg/types"

	"github.com/avodah-inc/provider-linear/config"
)

func main() {
	wd, _ := os.Getwd()
	root := wd
	if filepath.Base(wd) == "apis" {
		root = filepath.Dir(wd)
	}

	// Step 1: Upjet pipeline (namespaced scope)
	pc, err := config.GetProvider()
	if err != nil {
		panic(fmt.Sprintf("cannot build provider configuration: %v", err))
	}

	runner := &pipeline.PipelineRunner{
		DirAPIs:               filepath.Join(root, "apis"),
		DirControllers:        filepath.Join(root, "internal", "controller"),
		DirExamples:           filepath.Join(root, "examples-generated"),
		DirHack:               filepath.Join(root, "hack"),
		ModulePathAPIs:        filepath.Join(pc.ModulePath, "apis"),
		ModulePathControllers: filepath.Join(pc.ModulePath, "internal", "controller"),
		Scope:                 tjtypes.CRDScopeNamespaced,
	}
	runner.Run(pc)

	// Step 2: Post-generation patches
	runPostgen(root)

	// Step 3: controller-gen (object + crd)
	args := []string{"run", "sigs.k8s.io/controller-tools/cmd/controller-gen",
		"object:headerFile=" + filepath.Join(root, "hack", "boilerplate.go.txt"),
		"crd:allowDangerousTypes=true,crdVersions=v1",
		"paths=" + filepath.Join(root, "apis", "..."),
		"output:artifacts:config=" + filepath.Join(root, "package", "crds"),
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Sprintf("controller-gen failed: %v", err))
	}
}
