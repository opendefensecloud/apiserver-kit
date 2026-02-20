// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"go.opendefense.cloud/kit/apiserver/rest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mergeVersionedResourcesStorageMap", func() {
	It("should merge two empty maps", func() {
		a := map[string]map[string]rest.Storage{}
		b := map[string]map[string]rest.Storage{}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(BeEmpty())
	})

	It("should return clone of a when b is empty", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods":     nil,
				"services": nil,
			},
		}
		b := map[string]map[string]rest.Storage{}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveKey("v1"))
		Expect(result["v1"]).To(HaveLen(2))
		Expect(result["v1"]).To(HaveKey("pods"))
		Expect(result["v1"]).To(HaveKey("services"))
	})

	It("should return clone of b when a is empty", func() {
		a := map[string]map[string]rest.Storage{}
		b := map[string]map[string]rest.Storage{
			"v1beta1": {
				"orders": nil,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveKey("v1beta1"))
		Expect(result["v1beta1"]).To(HaveLen(1))
		Expect(result["v1beta1"]).To(HaveKey("orders"))
	})

	It("should merge resources from different versions", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods": nil,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1beta1": {
				"orders": nil,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveLen(2))
		Expect(result).To(HaveKey("v1"))
		Expect(result).To(HaveKey("v1beta1"))
		Expect(result["v1"]).To(HaveKey("pods"))
		Expect(result["v1beta1"]).To(HaveKey("orders"))
	})

	It("should merge resources from the same version", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods": nil,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1": {
				"services": nil,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveKey("v1"))
		Expect(result["v1"]).To(HaveLen(2))
		Expect(result["v1"]).To(HaveKey("pods"))
		Expect(result["v1"]).To(HaveKey("services"))
	})

	It("should preserve storage references from both maps", func() {
		storageA := &mockStorage{name: "storageA"}
		storageB := &mockStorage{name: "storageB"}
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods": storageA,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1": {
				"services": storageB,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result["v1"]["pods"]).To(Equal(storageA))
		Expect(result["v1"]["services"]).To(Equal(storageB))
	})

	It("should handle multiple versions and resources", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods":     nil,
				"services": nil,
			},
			"v1beta1": {
				"orders": nil,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1": {
				"nodes": nil,
			},
			"v1alpha1": {
				"fragments": nil,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveLen(3))
		Expect(result["v1"]).To(HaveLen(3))
		Expect(result["v1beta1"]).To(HaveLen(1))
		Expect(result["v1alpha1"]).To(HaveLen(1))
		Expect(result["v1"]).To(HaveKey("pods"))
		Expect(result["v1"]).To(HaveKey("services"))
		Expect(result["v1"]).To(HaveKey("nodes"))
		Expect(result["v1beta1"]).To(HaveKey("orders"))
		Expect(result["v1alpha1"]).To(HaveKey("fragments"))
	})

	It("should not modify input maps", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"pods": nil,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1": {
				"services": nil,
			},
		}
		// Keep copies to compare later
		aOriginal := map[string]map[string]rest.Storage{
			"v1": {
				"pods": nil,
			},
		}
		bOriginal := map[string]map[string]rest.Storage{
			"v1": {
				"services": nil,
			},
		}
		_ = mergeVersionedResourcesStorageMap(a, b)
		// Verify inputs weren't modified
		Expect(a).To(Equal(aOriginal))
		Expect(b).To(Equal(bOriginal))
	})

	It("should handle deeply nested structures", func() {
		a := map[string]map[string]rest.Storage{
			"v1": {
				"resource1": nil,
				"resource2": nil,
			},
			"v1beta1": {
				"resource3": nil,
			},
		}
		b := map[string]map[string]rest.Storage{
			"v1": {
				"resource4": nil,
			},
			"v1beta1": {
				"resource5": nil,
				"resource6": nil,
			},
			"v2": {
				"resource7": nil,
			},
		}
		result := mergeVersionedResourcesStorageMap(a, b)
		Expect(result).To(HaveLen(3))
		Expect(result["v1"]).To(HaveLen(3))
		Expect(result["v1beta1"]).To(HaveLen(3))
		Expect(result["v2"]).To(HaveLen(1))
	})
})

// mockStorage is a minimal implementation of rest.Storage for testing.
type mockStorage struct {
	name string
}

func (m *mockStorage) New() runtime.Object {
	return &mockStorage{name: m.name}
}

func (m *mockStorage) Destroy() {}

// DeepCopyObject implements runtime.Object interface.
func (m *mockStorage) DeepCopyObject() runtime.Object {
	if m == nil {
		return nil
	}
	clone := *m

	return &clone
}

// GetObjectKind implements runtime.Object interface.
func (m *mockStorage) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

var _ = Describe("Resource with interfaces", func() {
	Describe("Resource with SingularNameProvider", func() {
		It("should set singular name on the store", func() {
			obj := &mockResourceObject{
				gr:           schema.GroupResource{Group: "test.example.com", Resource: "testresources"},
				singularName: "testresource",
			}
			handler := Resource(obj, schema.GroupVersion{Group: "test.example.com", Version: "v1"})

			Expect(handler.groupVersions).To(HaveLen(1))
			Expect(handler.groupVersions[0]).To(Equal(schema.GroupVersion{Group: "test.example.com", Version: "v1"}))
		})
	})

	Describe("Resource with ShortNamesProvider", func() {
		It("should wrap store with ShortNamesProvider when short names provided", func() {
			obj := &mockResourceObject{
				gr:         schema.GroupResource{Group: "test.example.com", Resource: "testresources"},
				shortNames: []string{"tr", "tres"},
			}
			handler := Resource(obj, schema.GroupVersion{Group: "test.example.com", Version: "v1"})

			Expect(handler.groupVersions).To(HaveLen(1))
			Expect(handler.groupVersions[0]).To(Equal(schema.GroupVersion{Group: "test.example.com", Version: "v1"}))
		})
	})

	Describe("Resource with both SingularNameProvider and ShortNamesProvider", func() {
		It("should set both options correctly", func() {
			obj := &mockResourceObject{
				gr:           schema.GroupResource{Group: "test.example.com", Resource: "testresources"},
				singularName: "testresource",
				shortNames:   []string{"tr"},
			}
			handler := Resource(obj, schema.GroupVersion{Group: "test.example.com", Version: "v1"})

			Expect(handler.groupVersions).To(HaveLen(1))
			Expect(handler.groupVersions[0]).To(Equal(schema.GroupVersion{Group: "test.example.com", Version: "v1"}))
		})
	})

	Describe("Resource with no custom interfaces", func() {
		It("should work without implementing ShortNamesProvider or SingularNameProvider", func() {
			obj := &mockResourceObject{
				gr: schema.GroupResource{Group: "test.example.com", Resource: "testresources"},
			}
			handler := Resource(obj, schema.GroupVersion{Group: "test.example.com", Version: "v1"})

			Expect(handler.groupVersions).To(HaveLen(1))
		})
	})
})

type mockResourceObject struct {
	gr           schema.GroupResource
	singularName string
	shortNames   []string
}

func (m *mockResourceObject) GetObjectMeta() *metav1.ObjectMeta {
	return &metav1.ObjectMeta{}
}

func (m *mockResourceObject) NamespaceScoped() bool {
	return true
}

func (m *mockResourceObject) New() runtime.Object {
	return &mockResourceObject{}
}

func (m *mockResourceObject) NewList() runtime.Object {
	return &mockResourceList{}
}

func (m *mockResourceObject) GetGroupResource() schema.GroupResource {
	return m.gr
}

func (m *mockResourceObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (m *mockResourceObject) ShortNames() []string {
	return m.shortNames
}

func (m *mockResourceObject) GetSingularName() string {
	return m.singularName
}

func (m *mockResourceObject) DeepCopyInto(out *mockResourceObject) {
	*out = *m
}

func (m *mockResourceObject) DeepCopyObject() runtime.Object {
	if m == nil {
		return nil
	}
	outCopy := &mockResourceObject{}
	m.DeepCopyInto(outCopy)

	return outCopy
}

type mockResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []mockResourceObject
}

func (l *mockResourceList) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (l *mockResourceList) DeepCopyObject() runtime.Object {
	if l == nil {
		return nil
	}
	out := &mockResourceList{
		TypeMeta: l.TypeMeta,
		ListMeta: l.ListMeta,
	}
	if l.Items != nil {
		out.Items = make([]mockResourceObject, len(l.Items))
		copy(out.Items, l.Items)
	}

	return out
}
