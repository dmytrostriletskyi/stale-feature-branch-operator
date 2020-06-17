package stalefeaturebranch

import (
	"context"
	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

var isDebug = os.Getenv("IS_DEBUG")
var _ reconcile.Reconciler = &ReconcileStaleFeatureBranch{}

type ReconcileStaleFeatureBranch struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ReconcileStaleFeatureBranch) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var staleFeatureBranch featurebranchv1.StaleFeatureBranch

	if err := r.Client.Get(context.TODO(), request.NamespacedName, &staleFeatureBranch); err != nil {
		logger.Error(err, "Unable to fetch a stale feature branch.", request.NamespacedName)
		return reconcile.Result{}, nil
	}

	logger.Info(
		"Stale feature branch is being processing.",
		"namespaceSubstring", staleFeatureBranch.Spec.NamespaceSubstring,
		"afterDaysWithoutDeploy", staleFeatureBranch.Spec.AfterDaysWithoutDeploy,
		"checkEveryMinutes", staleFeatureBranch.Spec.CheckEveryMinutes,
		"isDebug", isDebug,
	)

	var allNamespaces corev1.NamespaceList

	if err := r.Client.List(context.TODO(), &allNamespaces); err != nil {
		logger.Error(err, "Unable to fetch the cluster's namespaces.")
		return reconcile.Result{}, nil
	}

	for _, namespace := range allNamespaces.Items {
		if r.IsNamespaceToBeDeleted(staleFeatureBranch, namespace) {
			logger.Info(
				"Namespace is being processing.",
				"namespaceName", namespace.Name,
				"namespaceCreationTimestamp", namespace.CreationTimestamp,
			)

			if err := r.Client.Delete(context.TODO(), &namespace); err != nil {
				logger.Error(err, "An error occurred while delete a namespace.", request.NamespacedName)
				return reconcile.Result{}, err
			}

			logger.Info("Namespace has been deleted.", "namespaceName", namespace.Name)
		}
	}

	requeueIn, _ := time.ParseDuration(strconv.Itoa(staleFeatureBranch.Spec.CheckEveryMinutes) + "m")
	return reconcile.Result{RequeueAfter: requeueIn}, nil
}

func (r *ReconcileStaleFeatureBranch) IsNamespaceToBeDeleted(staleFeatureBranch featurebranchv1.StaleFeatureBranch, namespace corev1.Namespace) bool {
	if !strings.Contains(namespace.Name, staleFeatureBranch.Spec.NamespaceSubstring) {
		return false
	}

	if InDebugMode == isDebug {
		logger.Info(
			"Namespace should be deleted due to debug mode is enabled.",
			"namespaceName", namespace.Name,
		)
		return true
	}

	differenceInDaysAsFloat := metav1.Now().Sub(namespace.CreationTimestamp.Time).Hours() / float64(HoursInDay)
	differenceInDaysAsInteger := int(differenceInDaysAsFloat)

	if differenceInDaysAsInteger > staleFeatureBranch.Spec.AfterDaysWithoutDeploy {
		return true
	}

	return false
}
