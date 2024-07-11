/*
Copyright 2024 openfluxcd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ArtifactKind is the string representation of an Artifact.
	ArtifactKind = "Artifact"
)

// ArtifactSpec defines the desired state of Artifact
type ArtifactSpec struct {
	// URL is the HTTP address of the Artifact as exposed by the controller
	// managing the Source. It can be used to retrieve the Artifact for
	// consumption, e.g. by another controller applying the Artifact contents.
	// +required
	URL string `json:"url"`

	// Revision is a human-readable identifier traceable in the origin source
	// system. It can be a Git commit SHA, Git tag, a Helm chart version, etc.
	// +required
	Revision string `json:"revision"`

	// Digest is the digest of the file in the form of '<algorithm>:<checksum>'.
	// +optional
	// +kubebuilder:validation:Pattern="^[a-z0-9]+(?:[.+_-][a-z0-9]+)*:[a-zA-Z0-9=_-]+$"
	Digest string `json:"digest,omitempty"`

	// LastUpdateTime is the timestamp corresponding to the last update of the
	// Artifact.
	// +required
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`

	// Size is the number of bytes in the file.
	// +optional
	Size *int64 `json:"size,omitempty"`

	// Metadata holds upstream information such as OCI annotations.
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ArtifactStatus defines the observed state of Artifact
type ArtifactStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Artifact is the Schema for the artifacts API
type Artifact struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArtifactSpec   `json:"spec,omitempty"`
	Status ArtifactStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ArtifactList contains a list of Artifact
type ArtifactList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Artifact `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Artifact{}, &ArtifactList{})
}

func (a *Artifact) GetArtifact() *sourcev1.Artifact {
	return &sourcev1.Artifact{
		Path:           "",
		URL:            a.Spec.URL,
		Revision:       a.Spec.Revision,
		Digest:         a.Spec.Digest,
		LastUpdateTime: a.Spec.LastUpdateTime,
		Size:           a.Spec.Size,
		Metadata:       a.Spec.Metadata,
	}
}
