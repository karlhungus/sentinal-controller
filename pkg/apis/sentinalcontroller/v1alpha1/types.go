/*
Copyright 2017 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SentinalDeployment is a specification for a SentinalDeployment resource
type SentinalDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SentinalDeploymentSpec   `json:"spec"`
	Status SentinalDeploymentStatus `json:"status"`
}

// SentinalDeploymentSpec is the spec for a Sentinal resource
type SentinalDeploymentSpec struct {
	StableDeploymentName  string      `json:"stableDeploymentName"`
	StableDesiredReplicas int32       `json:"stableReplicas"`
	CanaryDeploymentName  string      `json:"canaryDeploymentName"`
	CanaryContainers      []Container `json:"containers"`
	CanaryDesiredReplicas int32       `json:"canaryReplicas"`
}

// Container updating the Deployment Copy's containers with the canary image
type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// SentinalDeploymentStatus is the status for a SentinalDeployment resource
type SentinalDeploymentStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
	CanaryReplicas    int32 `json:"canaryReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SentinalDeploymentList is a list of SentinalDeployment resources
type SentinalDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SentinalDeployment `json:"items"`
}
