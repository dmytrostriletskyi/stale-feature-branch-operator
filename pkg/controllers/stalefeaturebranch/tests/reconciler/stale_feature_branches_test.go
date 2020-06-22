package reconciler

import (
	"context"
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/controllers/stalefeaturebranch"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"os"
	"testing"
	"time"

	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Case: delete stale feature branches.
// Expected: namespaces match specified namespace substring and older that specified days without deployed are deleted.
func TestReconcilerStaleFeatureBranches(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		oldNamespaceCreationTimestamp            = metav1.Date(
			2010, time.November, 10, 10, 10, 10, 10, time.UTC,
		)
		newNamespaceCreationTimestamp = metav1.Now()
		reconciler                    stalefeaturebranch.ReconcileStaleFeatureBranch
		request                       reconcile.Request
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

	firstNamespaceToBeDeleted := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: oldNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	secondNamespaceToBeDeleted := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-2",
			CreationTimestamp: oldNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	namespaceToNotBeDeleted := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project",
			CreationTimestamp: newNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		firstNamespaceToBeDeleted,
		secondNamespaceToBeDeleted,
		namespaceToNotBeDeleted,
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

	expectedNamespaceToNotBeDeleted := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project",
			CreationTimestamp: newNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

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
		1,
		len(allNamespaces.Items),
		"If 2 from 3 namespaces have namespace substring, they are deleted, the only 1 left.",
	)

	assert.Equal(
		t,
		expectedNamespaceToNotBeDeleted.Name,
		allNamespaces.Items[0].Name,
		"Expected namespace to not be deleted equals the single one in namespaces list.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}
