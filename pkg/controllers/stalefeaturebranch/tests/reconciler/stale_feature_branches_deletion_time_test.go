package reconciler

import (
	"context"
	"os"
	"testing"
	"time"

	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"

	"bou.ke/monkey"
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

// Case: delete stale feature branch after 1 day (24 hours) without deploy.
// Where: the only 23 hours and 59 minutes passed after its creation.
// Expected: namespace isn't deleted.
func TestReconcilerStaleFeatureBranchesDeletionTimeOneMinutesLess(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		namespaceCreationTimestamp               = metav1.Date(
			2010, time.January, 1, 0, 0, 0, 0, time.Local,
		)
		currentTimestamp = time.Date(
			2010, time.January, 1, 23, 59, 59, 59, time.Local,
		)
		reconciler stalefeaturebranch.ReconcileStaleFeatureBranch
		request    reconcile.Request
	)

	patch := monkey.Patch(time.Now, func() time.Time { return currentTimestamp })
	defer patch.Unpatch()

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

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: namespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		namespace,
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
		1,
		len(allNamespaces.Items),
		"Namespace isn't deleted as not enough time passed after its creation.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}

// Case: delete stale feature branch after 1 day (24 hours) without deploy.
// Where: 24 hours and 1 minute passed after its creation.
// Expected: namespace is deleted.
func TestReconcilerStaleFeatureBranchesDeletionTimeOneMinuteMore(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		namespaceCreationTimestamp               = metav1.Date(
			2010, time.January, 1, 0, 0, 0, 0, time.Local,
		)
		currentTimestamp = time.Date(
			2010, time.January, 2, 1, 0, 0, 0, time.Local,
		)
		reconciler stalefeaturebranch.ReconcileStaleFeatureBranch
		request    reconcile.Request
	)

	patch := monkey.Patch(time.Now, func() time.Time { return currentTimestamp })
	defer patch.Unpatch()

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

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: namespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		namespace,
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
		0,
		len(allNamespaces.Items),
		"Namespace isn't deleted as enough time passed after its creation..",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}

// Case: delete stale feature branch after 1 day (24 hours) without deploy.
// Where: exact 24 hours passed after its creation.
// Expected: namespace is deleted.
func TestReconcilerStaleFeatureBranchesDeletionTimeExactAfterDayesWithoutDeploy(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		namespaceCreationTimestamp               = metav1.Date(
			2010, time.January, 1, 0, 0, 0, 0, time.Local,
		)
		currentTimestamp = time.Date(
			2010, time.January, 2, 0, 0, 0, 0, time.Local,
		)
		reconciler stalefeaturebranch.ReconcileStaleFeatureBranch
		request    reconcile.Request
	)

	patch := monkey.Patch(time.Now, func() time.Time { return currentTimestamp })
	defer patch.Unpatch()

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

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: namespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		namespace,
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
		0,
		len(allNamespaces.Items),
		"Namespace is deleted as exact time passed after its creation.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}
