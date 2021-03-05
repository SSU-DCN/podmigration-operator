/*


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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodmigrationSpec defines the desired state of Podmigration
type PodmigrationSpec struct {

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas     int    `json:"replicas,omitempty"`
	SourcePod    string `json:"sourcePod,omitempty"`
	DestHost     string `json:"destHost,omitempty"`
	SnapshotPath string `json:"snapshotPath,omitempty"`

	// Label selector for pods. Existing ReplicaSets whose pods are
	// selected by this will be the ones affected by this deployment.
	// It must match the pod template's labels.
	Selector *metav1.LabelSelector  `json:"selector"`
	Template corev1.PodTemplateSpec `json:"template,omitempty"`

	// Template describes the pods that will be created.
	// +kubebuilder:validation:Required
	Action string `json:"action"`
	// ExcludeNode indicates a node that the Pod should not get scheduled on or get migrated
	// away from.
	// +kubebuilder:validation:Optional
	// ExcludeNodeSelector map[string]string `json:"excludeNodeSelector"`
}

// PodmigrationStatus defines the observed state of Podmigration
type PodmigrationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// State indicates the state of the MigratingPod
	State string `json:"state"`

	// CurrentRevision indicates the version of the MigratingPod to generate the current Pod
	CurrentRevision string `json:"currentRevision"`

	// ActivePod
	ActivePod string `json:"activePod"`
}

// +kubebuilder:object:root=true

// Podmigration is the Schema for the podmigrations API
type Podmigration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodmigrationSpec   `json:"spec,omitempty"`
	Status PodmigrationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodmigrationList contains a list of Podmigration
type PodmigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Podmigration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Podmigration{}, &PodmigrationList{})
}
