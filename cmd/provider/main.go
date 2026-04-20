// Package main is the entry point for provider-linear.
//
// It registers all Upjet-generated controllers and starts the Crossplane
// controller manager with the Terraform provider bridge running in no-fork
// (in-process) mode. The no-fork runtime embeds the Terraform provider
// directly in the Go binary, eliminating the need for a separate Terraform
// CLI binary or filesystem workspace.
//
// This file follows the standard Upjet provider main.go pattern. It will
// compile once the Upjet and Crossplane runtime dependencies are added to
// go.mod via `go mod tidy` after code generation.
package main

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/crossplane/crossplane-runtime/pkg/feature"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	tjcontroller "github.com/crossplane/upjet/v2/pkg/controller"

	"github.com/avodah-inc/provider-linear/apis"
	"github.com/avodah-inc/provider-linear/config"
	"github.com/avodah-inc/provider-linear/internal/controller"
	"github.com/avodah-inc/provider-linear/internal/features"
)

func main() {
	var (
		app = kingpin.New(filepath.Base(os.Args[0]),
			"Crossplane provider for Linear (terraform-community-providers/linear)").
			DefaultEnvars()

		debug = app.Flag("debug",
			"Run with debug logging.").
			Short('d').Bool()

		syncInterval = app.Flag("sync",
			"How often all resources will be double-checked for drift from the desired state.").
			Short('s').Default("1h").Duration()

		pollInterval = app.Flag("poll",
			"How often individual resources will be checked for drift from the desired state.").
			Default("10m").Duration()

		leaderElection = app.Flag("leader-election",
			"Use leader election for the controller manager.").
			Short('l').Default("false").Envar("LEADER_ELECTION").Bool()

		maxReconcileRate = app.Flag("max-reconcile-rate",
			"The global maximum rate per second at which resources may be checked for drift from the desired state.").
			Default("10").Int()

		enableManagementPolicies = app.Flag("enable-management-policies",
			"Enable support for Management Policies.").
			Default("true").Envar("ENABLE_MANAGEMENT_POLICIES").Bool()
	)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-linear"))
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only use it when running
		// in debug mode.
		ctrl.SetLogger(zl)
	}

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Cache: cache.Options{
			SyncPeriod: syncInterval,
		},
		LeaderElection:             *leaderElection,
		LeaderElectionID:           "crossplane-leader-election-provider-linear",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	// Configure feature flags for the Crossplane runtime.
	featureFlags := &feature.Flags{}
	if *enableManagementPolicies {
		featureFlags.Enable(features.EnableAlphaManagementPolicies)
		log.Info("Alpha feature enabled", "flag", features.EnableAlphaManagementPolicies)
	}

	// Build the Upjet provider configuration. This bridges the upstream
	// Terraform provider (terraform-community-providers/linear) into
	// Crossplane controllers and CRDs via no-fork mode.
	provider, err := config.GetProvider()
	kingpin.FatalIfError(err, "Cannot build Upjet provider configuration")

	// Configure the Upjet controller options. The no-fork runtime mode
	// means the Terraform provider runs in-process — no external Terraform
	// binary or workspace directory is needed.
	o := tjcontroller.Options{
		Provider: provider,
		// TODO(task 6): Wire SetupFn to the auth module's
		// TerraformSetupBuilder once authentication is implemented.
		// SetupFn configures the Terraform provider with credentials
		// from the ProviderConfig during each reconciliation cycle.
	}
	o.Logger = log
	o.GlobalRateLimiter = ratelimiter.NewGlobal(*maxReconcileRate)
	o.PollInterval = *pollInterval
	o.MaxConcurrentReconciles = *maxReconcileRate
	o.Features = featureFlags

	// Register all Linear API types with the controller manager scheme.
	kingpin.FatalIfError(apis.AddToScheme(mgr.GetScheme()), "Cannot add Linear APIs to scheme")

	// Register all Upjet-generated controllers with the manager.
	// After `make generate`, controller.Setup registers a reconciler for
	// each managed resource: Team, TeamLabel, TeamWorkflow, Template,
	// WorkflowState, WorkspaceLabel, WorkspaceSettings, and Workspace.
	kingpin.FatalIfError(controller.Setup(mgr, o), "Cannot setup Linear controllers")

	// Start the controller manager. This blocks until the context is
	// cancelled (e.g., SIGTERM).
	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}

// Ensure time is used (referenced by Duration flags).
var _ time.Duration
