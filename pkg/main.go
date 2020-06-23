package main

import (
	"context"
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis"
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/controllers"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
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
