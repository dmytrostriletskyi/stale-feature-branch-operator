package stalefeaturebranch

import (
	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var logger = logf.Log.WithName("stale-feature-branch-controller")

func CreateController(mgr manager.Manager, r reconcile.Reconciler) error {

	c, err := controller.New("stalefeaturebranch-controller", mgr, controller.Options{Reconciler: r})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &featurebranchv1.StaleFeatureBranch{}}, &handler.EnqueueRequestForObject{})

	if err != nil {
		return err
	}

	return nil
}
