module github.com/dmytrostriletskyi/stale-feature-branch-operator

go 1.13

require (
	bou.ke/monkey v1.0.2
	github.com/operator-framework/operator-sdk v0.18.1
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
