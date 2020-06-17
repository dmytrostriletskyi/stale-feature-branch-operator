package apis

import (
	"github.com/dmytrostriletskyi/stale-feature-branch-operator/pkg/apis/featurebranch/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var SchemeBuilder runtime.SchemeBuilder

func RegisterSchemes(scheme *runtime.Scheme) error {
	SchemeBuilder = append(SchemeBuilder, v1.SchemeBuilder.AddToScheme)

	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}
