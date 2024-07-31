package testutils

import (
	"context"
	"fmt"
	"github.com/fluxcd/pkg/testserver"
	"github.com/opencontainers/go-digest"
	"github.com/openfluxcd/artifact/api/commonv1"
	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func ApplyArtifact(client ctrl.Client, testServer *testserver.ArtifactServer, artifact ctrl.ObjectKey, urlpath string, revision string) (*commonv1.SourceRef, error) {
	b, _ := os.ReadFile(filepath.Join(testServer.Root(), urlpath))
	dig := digest.SHA256.FromBytes(b)

	url := fmt.Sprintf("%s/%s", testServer.URL(), urlpath)

	art := &artifactv1.Artifact{
		TypeMeta: metav1.TypeMeta{
			Kind:       artifactv1.ArtifactKind,
			APIVersion: artifactv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      artifact.Name,
			Namespace: artifact.Namespace,
		},
		Spec: artifactv1.ArtifactSpec{
			URL:            url,
			Revision:       revision,
			Digest:         dig.String(),
			LastUpdateTime: metav1.Now(),
		},
	}

	sourceref := &commonv1.SourceRef{
		APIVersion: art.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       art.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  art.GetNamespace(),
		Name:       art.GetName(),
	}

	opt := []ctrl.PatchOption{
		ctrl.ForceOwnership,
		ctrl.FieldOwner("kustomize-controller"),
	}

	if err := client.Patch(context.Background(), art, ctrl.Apply, opt...); err != nil {
		return nil, err
	}

	return sourceref, nil
}

func ApplyGenericSource(client ctrl.Client, testServer *testserver.ArtifactServer, source ctrl.Object, urlpath string, revision string) (*commonv1.SourceRef, error) {
	if source.GetObjectKind().GroupVersionKind().Version == "" || source.GetObjectKind().GroupVersionKind().Kind == "" {
		return nil, fmt.Errorf("source APIVersion and Kind must be set")
	}
	sourceCopy := source.DeepCopyObject().(ctrl.Object)
	if err := client.Create(context.Background(), sourceCopy); err != nil {
		return nil, err
	}
	source.SetUID(sourceCopy.GetUID())

	sourceref := &commonv1.SourceRef{
		APIVersion: source.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       source.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  source.GetNamespace(),
		Name:       source.GetName(),
	}

	b, _ := os.ReadFile(filepath.Join(testServer.Root(), urlpath))
	dig := digest.SHA256.FromBytes(b)

	url := fmt.Sprintf("%s/%s", testServer.URL(), urlpath)

	art := &artifactv1.Artifact{
		TypeMeta: metav1.TypeMeta{
			Kind:       artifactv1.ArtifactKind,
			APIVersion: artifactv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      source.GetName(),
			Namespace: source.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: source.GetObjectKind().GroupVersionKind().GroupVersion().String(),
					Kind:       source.GetObjectKind().GroupVersionKind().Kind,
					Name:       source.GetName(),
					UID:        source.GetUID(),
				},
			},
		},
		Spec: artifactv1.ArtifactSpec{
			URL:            url,
			Revision:       revision,
			Digest:         dig.String(),
			LastUpdateTime: metav1.Now(),
		},
	}

	opt := []ctrl.PatchOption{
		ctrl.ForceOwnership,
		ctrl.FieldOwner("kustomize-controller"),
	}

	if err := client.Patch(context.Background(), art, ctrl.Apply, opt...); err != nil {
		return nil, err
	}

	return sourceref, nil
}
