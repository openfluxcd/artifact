package utils

import (
	"fmt"
	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func OwnerReferenceIndex() func(o ctrlclient.Object) []string {
	return func(o ctrlclient.Object) []string {
		keys := []string{fmt.Sprintf("%s/%s/%s/%s", artifactv1.GroupVersion.Group, artifactv1.ArtifactKind, o.GetNamespace(), o.GetName())}
		for _, ref := range o.GetOwnerReferences() {
			keys = append(keys, fmt.Sprintf("%s/%s/%s/%s", ExtractGroupName(ref.APIVersion), ref.Kind, o.GetNamespace(), ref.Name))
		}
		return keys
	}
}
