package commonv1

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/openfluxcd/artifact/utils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CrossNamespaceSourceReference contains enough information to let you locate the
// typed Kubernetes resource object at cluster level.
type SourceRef struct {
	// API version of the referent.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referent.
	// +required
	Kind string `json:"kind"`

	// Name of the referent.
	// +required
	Name string `json:"name"`

	// Namespace of the referent, defaults to the namespace of the Kubernetes
	// resource object that contains the reference.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func (s *SourceRef) GetObjectKey() ctrlclient.ObjectKey {
	return ctrlclient.ObjectKey{
		Namespace: s.Namespace,
		Name:      s.Name,
	}
}

func (s *SourceRef) GetGroupKind() schema.GroupKind {
	if s.APIVersion == "" {
		return schema.GroupKind{
			Group: sourcev1.GroupVersion.Group,
			Kind:  s.Kind,
		}
	}

	return schema.GroupKind{
		Group: utils.ExtractGroupName(s.APIVersion),
		Kind:  s.Kind,
	}
}

func (s *SourceRef) GetName() string {
	return s.Name
}

func (s *SourceRef) GetNamespace() string {
	return s.Namespace
}
