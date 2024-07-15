package action

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type SourceRevisionChangePredicate struct {
	predicate.Funcs
}

func (SourceRevisionChangePredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}

	oldSource, ok := e.ObjectOld.(ArtifactSource)
	if !ok {
		return false
	}

	newSource, ok := e.ObjectNew.(ArtifactSource)
	if !ok {
		return false
	}

	if oldSource.GetArtifact() == nil && newSource.GetArtifact() != nil {
		return true
	}

	if oldSource.GetArtifact() != nil && newSource.GetArtifact() != nil &&
		!oldSource.GetArtifact().HasRevision(newSource.GetArtifact().Revision) {
		return true
	}

	return false
}
