package matchers

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1b2 "github.com/fluxcd/source-controller/api/v1beta2"
	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MapMatcher map[schema.GroupKind]client.Object

func (m MapMatcher) Match(gk schema.GroupKind) bool {
	_, ok := m[gk]
	return ok
}

func (m MapMatcher) Create(gk schema.GroupKind) client.Object {
	obj := m[gk]
	if obj == nil {
		return nil
	}
	return obj.DeepCopyObject().(client.Object)
}

var (
	BuiltinGeneralSourceKinds = MapMatcher{
		{sourcev1.GroupVersion.Group, sourcev1.GitRepositoryKind}:     &sourcev1.GitRepository{},
		{sourcev1b2.GroupVersion.Group, sourcev1b2.BucketKind}:        &sourcev1b2.Bucket{},
		{sourcev1b2.GroupVersion.Group, sourcev1b2.OCIRepositoryKind}: &sourcev1b2.OCIRepository{},
	}

	BuiltinFluxSourceKinds = MapMatcher{
		{sourcev1.GroupVersion.Group, sourcev1.GitRepositoryKind}:     &sourcev1.GitRepository{},
		{sourcev1b2.GroupVersion.Group, sourcev1b2.BucketKind}:        &sourcev1b2.Bucket{},
		{sourcev1b2.GroupVersion.Group, sourcev1b2.OCIRepositoryKind}: &sourcev1b2.OCIRepository{},
		{sourcev1.GroupVersion.Group, sourcev1.HelmRepositoryKind}:    &sourcev1.HelmRepository{},
		{sourcev1.GroupVersion.Group, sourcev1.HelmChartKind}:         &sourcev1.HelmChart{},
		{artifactv1.GroupVersion.Group, artifactv1.ArtifactKind}:      &artifactv1.Artifact{},
	}

	BuiltinHelmIndexSourceKinds = MapMatcher{
		{sourcev1.GroupVersion.Group, sourcev1.HelmRepositoryKind}: &sourcev1.HelmRepository{},
	}

	DynamicSourceKinds = Not(BuiltinFluxSourceKinds)

	AllSourceKinds = MatchFunc(func(gk schema.GroupKind) bool {
		return true
	})
)

type SourceMatcher interface {
	Match(gk schema.GroupKind) bool
}

type MatchFunc func(gk schema.GroupKind) bool

func (m MatchFunc) Match(gk schema.GroupKind) bool {
	return m(gk)
}

func And(matchers ...SourceMatcher) SourceMatcher {
	return MatchFunc(func(gk schema.GroupKind) bool {
		for _, m := range matchers {
			if !m.Match(gk) {
				return false
			}
		}
		return true
	})
}

func Or(matchers ...SourceMatcher) SourceMatcher {
	return MatchFunc(func(gk schema.GroupKind) bool {
		for _, m := range matchers {
			if m.Match(gk) {
				return true
			}
		}
		return false
	})
}

func Not(matcher SourceMatcher) SourceMatcher {
	return MatchFunc(func(gk schema.GroupKind) bool {
		return !matcher.Match(gk)
	})
}
