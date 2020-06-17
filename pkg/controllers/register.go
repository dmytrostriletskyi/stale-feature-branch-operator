package controllers

import (
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/controllers/stalefeaturebranch"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func RegisterControllers(manager manager.Manager) error {
	staleFeatureBranchReconcile := &stalefeaturebranch.ReconcileStaleFeatureBranch{
		Client: manager.GetClient(),
		Scheme: manager.GetScheme(),
	}

	if err := stalefeaturebranch.CreateController(manager, staleFeatureBranchReconcile); err != nil {
		return err
	}

	return nil
}
