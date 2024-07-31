package action

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func DefaultRequestMapper(list []runtime.Object) []reconcile.Request {
	var requests []reconcile.Request
	for _, obj := range list {
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(obj.(client.Object))})
	}
	return requests
}

func TriggerAlwaysPredicate(_ ActionResource, _ ArtifactSource) bool {
	return true
}
