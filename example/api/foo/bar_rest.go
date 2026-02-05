// Copyright 2026 BWI GmbH and contributors
// SPDX-License-Identifier: Apache-2.0

package foo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"go.opendefense.cloud/kit/apiserver/resource"
)

var _ resource.Object = &Bar{}

func (o *Bar) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

func (o *Bar) NamespaceScoped() bool {
	return true
}

func (o *Bar) New() runtime.Object {
	return &Bar{}
}

func (o *Bar) NewList() runtime.Object {
	return &BarList{}
}

func (o *Bar) GetGroupResource() schema.GroupResource {
	return SchemeGroupVersion.WithResource("bars").GroupResource()
}

var _ resource.Object = &ClusterBar{}

func (o *ClusterBar) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

func (o *ClusterBar) NamespaceScoped() bool {
	return false
}

func (o *ClusterBar) New() runtime.Object {
	return &ClusterBar{}
}

func (o *ClusterBar) NewList() runtime.Object {
	return &ClusterBarList{}
}

func (o *ClusterBar) GetGroupResource() schema.GroupResource {
	return SchemeGroupVersion.WithResource("clusterbars").GroupResource()
}
