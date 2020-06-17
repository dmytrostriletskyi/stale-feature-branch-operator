package main

import (
	"context"
	"flag"
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis"
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/controllers"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var logger = logf.Log.WithName("main")

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	logger.Info(
		"System information is fetched.",
		"Operator version", GetProjectVersion(),
		"Go version", runtime.Version(),
		"Go operating system", runtime.GOOS,
		"Go architecture", runtime.GOARCH,
		"Operator SDK version", sdkVersion.Version,
	)

	cfg, err := config.GetConfig()

	if err != nil {
		logger.Error(err, "Error occurred while getting configurations.")
		os.Exit(FailedExitCode)
	}

	err = leader.Become(context.TODO(), "stale-feature-branch-operator-lock")

	if err != nil {
		logger.Error(err, "Error occurred while becoming a leader.")
		os.Exit(FailedExitCode)
	}

	mgr, err := manager.New(cfg, manager.Options{
		Namespace: WatchAllNamespaces,
	})

	if err != nil {
		logger.Error(err, "Error occurred while initialization of manager.")
		os.Exit(FailedExitCode)
	}

	if err := apis.RegisterSchemes(mgr.GetScheme()); err != nil {
		logger.Error(err, "Error occurred while registering schemes.")
		os.Exit(FailedExitCode)
	}

	if err := controllers.RegisterControllers(mgr); err != nil {
		logger.Error(err, "Error occurred while registering controllers.")
		os.Exit(FailedExitCode)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "Manager exited with error.")
		os.Exit(FailedExitCode)
	}
}
