package reconciler

import (
	"context"
	"os"
	"testing"

	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"

	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/controllers/stalefeaturebranch"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Case: delete new stale feature branches.
// Expected: as all namespaces're are new, they aren't deleted.
func TestReconcilerNewStaleFeatureBranches(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		newNamespaceCreationTimestamp            = metav1.Now()
		reconciler                               stalefeaturebranch.ReconcileStaleFeatureBranch
		request                                  reconcile.Request
	)

	if err := os.Setenv("IS_DEBUG", "false"); err != nil {
		t.Fatalf("An error occurred while enabling debug: (%v)", err)
	}

	staleFeatureBranch := &featurebranchv1.StaleFeatureBranch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      staleFeatureBranchName,
			Namespace: staleFeatureBranchNamespace,
		},
		Spec: featurebranchv1.StaleFeatureBranchSpec{
			NamespaceSubstring:     staleFeatureBranchNamespaceSubstring,
			AfterDaysWithoutDeploy: staleFeatureBranchAfterDaysWithoutDeploy,
			CheckEveryMinutes:      staleFeatureBranchCheckEveryMinutes,
		},
	}

	firstNewNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: newNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	secondNewNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-2",
			CreationTimestamp: newNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		firstNewNamespace,
		secondNewNamespace,
	}

	s := scheme.Scheme
	s.AddKnownTypes(featurebranchv1.SchemeGroupVersion, staleFeatureBranch)

	reconciler = stalefeaturebranch.ReconcileStaleFeatureBranch{
		Client: fake.NewFakeClientWithScheme(s, objects...),
		Scheme: s,
	}

	request = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      staleFeatureBranchName,
			Namespace: staleFeatureBranchNamespace,
		},
	}

	// Testing.

	res, err := reconciler.Reconcile(request)

	if err != nil {
		t.Fatalf("An error occurred while calling the reconcile with a request: (%v)", err)
	}

	var allNamespaces corev1.NamespaceList

	if err := reconciler.Client.List(context.TODO(), &allNamespaces); err != nil {
		t.Fatalf("An error occurred while fetching all namespaces: (%v)", err)
	}

	assert.Equal(
		t,
		2,
		len(allNamespaces.Items),
		"As all namespaces're are new, they aren't deleted.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}
