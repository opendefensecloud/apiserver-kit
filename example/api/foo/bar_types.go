// Copyright 2026 BWI GmbH and contributors
// SPDX-License-Identifier: Apache-2.0

package foo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BarSpec struct {
	Message string `json:"message"`
}

type BarStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bar is just an example.
type Bar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   BarSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status BarStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BarList is a list of Bar objects.
type BarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []Bar `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterBar is the Schema for the endpoints API
type ClusterBar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   BarSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status BarStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterBarList is a list of Bar objects.
type ClusterBarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []ClusterBar `json:"items" protobuf:"bytes,2,rep,name=items"`
}
