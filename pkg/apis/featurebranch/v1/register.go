// +k8s:deepcopy-gen=package,register
// +groupName=feature-branch.dmytrostriletskyi.com
package v1

import (
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: featurebranch.ApiGroupName, Version: featurebranch.ApiGroupVersion}
	SchemeBuilder      = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
