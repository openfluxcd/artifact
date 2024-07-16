package utils

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type SourceRefProvider interface {
	GetObjectKey() ctrlclient.ObjectKey
	GetGroupKind() schema.GroupKind
	GetName() string
	GetNamespace() string
	String() string
}

func NewSourceRef(g, k, ns, name string) SourceRefProvider {
	return &DefaultSourceRef{
		GroupKind: schema.GroupKind{
			Group: g,
			Kind:  k,
		},
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
	}
}

type DefaultSourceRef struct {
	schema.GroupKind
	types.NamespacedName
}

func (d *DefaultSourceRef) GetObjectKey() ctrlclient.ObjectKey {
	return d.NamespacedName
}

func (d *DefaultSourceRef) GetGroupKind() schema.GroupKind {
	return d.GroupKind
}

func (d *DefaultSourceRef) GetName() string {
	return d.Name
}

func (d *DefaultSourceRef) GetNamespace() string {
	return d.Namespace
}

func (d *DefaultSourceRef) String() string {
	if d.GetNamespace() != "" {
		return fmt.Sprintf("%s/%s/%s/%s", d.GetGroupKind().Group, d.GetGroupKind().Kind, d.GetNamespace(), d.GetName())
	}
	return fmt.Sprintf("%s/%s/%s", d.GetGroupKind().Group, d.GetGroupKind().Kind, d.GetName())
}

func NormalizedSourceRef(ref SourceRefProvider, defns string) SourceRefProvider {
	if ref.GetNamespace() == "" {
		return NewSourceRef(ref.GetGroupKind().Group, ref.GetGroupKind().Kind, defns, ref.GetName())
	}
	return ref
}

func KeyForReference(o metav1.Object, ref SourceRefProvider) string {
	gk := ref.GetGroupKind()

	if gk.Kind == "" {
		return ""
	}

	namespace := o.GetNamespace()
	if ref.GetNamespace() != "" {
		namespace = ref.GetNamespace()
	}
	return fmt.Sprintf("%s/%s/%s/%s", gk.Group, gk.Kind, namespace, ref.GetName())
}

func ExtractGroupName(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}
