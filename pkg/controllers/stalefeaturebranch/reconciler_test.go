package stalefeaturebranch

import (
	"bou.ke/monkey"
	"context"
	"os"
	"testing"
	"time"

	featurebranchv1 "github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Case: delete stale feature branches.
// Where: some namespaces aren't feature branch namespaces.
// Expected: namespaces match namespace substring and older that days without deployed are deleted.
func TestReconcilerStaleFeatureBranches(t *testing.T) {
	// Set up data for tests.
	var (
		notFeatureBranchNamespaceName			 = "project"
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		oldNamespaceCreationTimestamp            = metav1.Date(
			2010, time.November, 10, 10, 10, 10, 10, time.UTC,
		)
		currentTimestamp = metav1.Now()
		reconciler       ReconcileStaleFeatureBranch
		request          reconcile.Request
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

	firstFeatureBranchNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-1",
			CreationTimestamp: oldNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	secondFeatureBranchNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "project-pr-2",
			CreationTimestamp: oldNamespaceCreationTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	notFeatureBranchNamespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              notFeatureBranchNamespaceName,
			CreationTimestamp: currentTimestamp,
		},
		Spec:   corev1.NamespaceSpec{},
		Status: corev1.NamespaceStatus{},
	}

	objects := []runtime.Object{
		staleFeatureBranch,
		firstFeatureBranchNamespace,
		secondFeatureBranchNamespace,
		notFeatureBranchNamespace,
	}

	s := scheme.Scheme
	s.AddKnownTypes(featurebranchv1.SchemeGroupVersion, staleFeatureBranch)

	reconciler = ReconcileStaleFeatureBranch{
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
			Name:              notFeatureBranchNamespaceName,
			CreationTimestamp: currentTimestamp,
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
		"2 from 3 namespaces have namespace substring, they are deleted, the only 1 left.",
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

// Case: delete stale feature branches.
// Where: all namespaces are new namespaces.
// Expected: namespaces aren't deleted.
func TestReconcilerNewStaleFeatureBranches(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		newNamespaceCreationTimestamp            = metav1.Now()
		reconciler                               ReconcileStaleFeatureBranch
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

	reconciler = ReconcileStaleFeatureBranch{
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
		"As all namespaces're new, they aren't deleted.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}


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
		reconciler ReconcileStaleFeatureBranch
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

	reconciler = ReconcileStaleFeatureBranch{
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
		reconciler ReconcileStaleFeatureBranch
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

	reconciler = ReconcileStaleFeatureBranch{
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
		reconciler ReconcileStaleFeatureBranch
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

	reconciler = ReconcileStaleFeatureBranch{
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


// Case: delete new stale feature branches.
// Where: debug is enabled.
// Expected: new namespaces are deleted.
func TestReconcilerStaleFeatureBranchesDebugEnabled(t *testing.T) {
	// Set up data for tests.
	var (
		staleFeatureBranchName                   = "stale-feature-branch-operator"
		staleFeatureBranchNamespace              = "stale-feature-branch-operator"
		staleFeatureBranchNamespaceSubstring     = "-pr-"
		staleFeatureBranchAfterDaysWithoutDeploy = 1
		staleFeatureBranchCheckEveryMinutes      = 1
		newNamespaceCreationTimestamp            = metav1.Now()
		reconciler                               ReconcileStaleFeatureBranch
		request                                  reconcile.Request
	)

	if err := os.Setenv("IS_DEBUG", "true"); err != nil {
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

	reconciler = ReconcileStaleFeatureBranch{
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
		"As debug is enabled, all namespaces are deleted not looking for oldness.",
	)

	assert.Equal(
		t,
		float64(staleFeatureBranchCheckEveryMinutes),
		res.RequeueAfter.Minutes(),
		"Check every minutes parameter equals the reconcile's requeue after one.",
	)
}
