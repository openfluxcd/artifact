package utils

import (
	"fmt"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func OwnerReferenceIndex() func(o ctrlclient.Object) []string {
	return func(o ctrlclient.Object) []string {
		gvk := o.GetObjectKind().GroupVersionKind()
		keys := []string{fmt.Sprintf("%s/%s/%s/%s", gvk.Group, gvk.Kind, o.GetNamespace(), o.GetName())}
		for _, ref := range o.GetOwnerReferences() {
			keys = append(keys, fmt.Sprintf("%s/%s/%s/%s", ExtractGroupName(ref.APIVersion), ref.Kind, o.GetNamespace(), ref.Name))
		}
		return keys
	}
}
