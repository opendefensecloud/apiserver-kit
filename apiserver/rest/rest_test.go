// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeRESTOptionsGetter returns a fixed RESTOptions, standing in for the
// storage-backed getter in unit tests.
type fakeRESTOptionsGetter struct {
	prefix string
}

func (f fakeRESTOptionsGetter) GetRESTOptions(_ schema.GroupResource, _ runtime.Object) (generic.RESTOptions, error) {
	return generic.RESTOptions{ResourcePrefix: f.prefix}, nil
}

var _ = Describe("Group-scoped storage keys", func() {
	It("scopes the ResourcePrefix by group and lowercases the resource", func() {
		Expect(GroupScopedResourcePrefix(schema.GroupResource{Group: "net.example.com", Resource: "Widgets"})).
			To(Equal("net.example.com/widgets"))
	})

	It("gives same-named resources in different groups distinct keys", func() {
		a := GroupScopedResourcePrefix(schema.GroupResource{Group: "net.example.com", Resource: "widgets"})
		b := GroupScopedResourcePrefix(schema.GroupResource{Group: "platform.example.com", Resource: "widgets"})
		Expect(a).ToNot(Equal(b))
	})

	It("overrides the delegate ResourcePrefix so the group is encoded in the key", func() {
		gr := schema.GroupResource{Group: "net.example.com", Resource: "widgets"}
		getter := groupScopedRESTOptionsGetter{
			delegate:       fakeRESTOptionsGetter{prefix: "widgets"},
			resourcePrefix: GroupScopedResourcePrefix(gr),
		}
		opts, err := getter.GetRESTOptions(gr, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(opts.ResourcePrefix).To(Equal("net.example.com/widgets"))
	})
})

var _ = Describe("GetAttrs and SelectableFields", func() {
	It("should extract labels and fields from a resource.Object", func() {
		obj := &testObj{}
		obj.SetLabels(map[string]string{"foo": "bar"})
		obj.Name = "myname"
		obj.Namespace = "ns"
		labelsSet, fieldsSet, err := GetAttrs(obj)
		Expect(err).ToNot(HaveOccurred())
		Expect(labelsSet).To(HaveKeyWithValue("foo", "bar"))
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.name", "myname"))
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.namespace", "ns"))
	})

	It("SelectableFields should return correct fields from ObjectMeta", func() {
		meta := &metav1.ObjectMeta{Name: "n", Namespace: "ns", Labels: map[string]string{"x": "y"}}
		fieldsSet := SelectableFields(meta)
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.name", "n"))
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.namespace", "ns"))
	})
})

// selectableObj implements SelectableFieldsProvider and
// SupportedFieldSelectorsProvider, exposing a spec.region field selector.
type selectableObj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Region string
}

func (t *selectableObj) DeepCopyObject() runtime.Object {
	if t == nil {
		return nil
	}
	clone := *t

	return &clone
}

func (t *selectableObj) GetObjectMeta() *metav1.ObjectMeta { return &t.ObjectMeta }
func (t *selectableObj) NamespaceScoped() bool             { return false }
func (t *selectableObj) New() runtime.Object               { return &selectableObj{} }
func (t *selectableObj) NewList() runtime.Object           { return nil }

func (t *selectableObj) GetGroupResource() schema.GroupResource {
	return schema.GroupResource{Group: "arc", Resource: "selectableobjs"}
}

// SelectableFields implements SelectableFieldsProvider.
func (t *selectableObj) SelectableFields() fields.Set {
	return fields.Set{"spec.region": t.Region}
}

// SupportedFieldSelectors implements SupportedFieldSelectorsProvider.
func (t *selectableObj) SupportedFieldSelectors() []string {
	return []string{"spec.region"}
}

// derivedObj implements only SelectableFieldsProvider, so its advertised
// field-selector keys are derived from the emitted fields.
type derivedObj struct {
	metav1.ObjectMeta
}

func (t *derivedObj) SelectableFields() fields.Set {
	return fields.Set{"spec.zone": "", "spec.region": ""}
}

// overrideObj implements both interfaces; SupportedFieldSelectors takes
// precedence over the keys emitted by SelectableFields.
type overrideObj struct {
	metav1.ObjectMeta
}

func (t *overrideObj) SelectableFields() fields.Set {
	return fields.Set{"spec.region": ""}
}

func (t *overrideObj) SupportedFieldSelectors() []string {
	return []string{"spec.legacyHost"}
}

// reservedKeyObj returns a reserved ObjectMeta key ("metadata.name") alongside a
// spec key, to verify provider fields cannot overwrite ObjectMeta-derived keys.
type reservedKeyObj struct {
	selectableObj
}

func (t *reservedKeyObj) SelectableFields() fields.Set {
	return fields.Set{"metadata.name": "spoofed", "spec.region": "eu"}
}

var _ = Describe("FieldSelectorKeys", func() {
	It("returns nil for a resource that advertises no extra selectors", func() {
		Expect(FieldSelectorKeys(&testObj{})).To(BeNil())
	})

	It("derives sorted keys from SelectableFieldsProvider when Supported is absent", func() {
		Expect(FieldSelectorKeys(&derivedObj{})).To(Equal([]string{"spec.region", "spec.zone"}))
	})

	It("prefers SupportedFieldSelectorsProvider over the derived default", func() {
		Expect(FieldSelectorKeys(&overrideObj{})).To(Equal([]string{"spec.legacyHost"}))
	})
})

var _ = Describe("Selectable spec fields", func() {
	It("GetAttrs merges SelectableFieldsProvider fields on top of ObjectMeta fields", func() {
		obj := &selectableObj{Region: "eu"}
		obj.Name = "a"
		_, fieldsSet, err := GetAttrs(obj)
		Expect(err).ToNot(HaveOccurred())
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.name", "a"))
		Expect(fieldsSet).To(HaveKeyWithValue("spec.region", "eu"))
	})

	It("does not let provider fields overwrite reserved ObjectMeta keys", func() {
		obj := &reservedKeyObj{}
		obj.Name = "real"
		_, fieldsSet, err := GetAttrs(obj)
		Expect(err).ToNot(HaveOccurred())
		// The reserved metadata.name keeps the real ObjectMeta value, not the
		// provider's spoofed one; additive spec keys are still merged.
		Expect(fieldsSet).To(HaveKeyWithValue("metadata.name", "real"))
		Expect(fieldsSet).To(HaveKeyWithValue("spec.region", "eu"))
	})

	It("a SelectionPredicate matches only objects with the requested spec field value", func() {
		sel, err := fields.ParseSelector("spec.region=eu")
		Expect(err).ToNot(HaveOccurred())
		pred := DefaultStrategy{}.Match(labels.Everything(), sel)

		euObj := &selectableObj{Region: "eu"}
		euObj.Name = "a"
		usObj := &selectableObj{Region: "us"}
		usObj.Name = "b"

		matchesEu, err := pred.Matches(euObj)
		Expect(err).ToNot(HaveOccurred())
		Expect(matchesEu).To(BeTrue())

		matchesUs, err := pred.Matches(usObj)
		Expect(err).ToNot(HaveOccurred())
		Expect(matchesUs).To(BeFalse())
	})

	It("RegisterFieldLabelConversions accepts supported keys and rejects unknown ones", func() {
		scheme := runtime.NewScheme()
		gvk := schema.GroupVersionKind{Group: "arc", Version: "v1alpha1", Kind: "SelectableObj"}
		Expect(RegisterFieldLabelConversions(scheme, gvk, []string{"spec.region"})).To(Succeed())

		// Supported spec key passes through unchanged.
		label, value, err := scheme.ConvertFieldLabel(gvk, "spec.region", "eu")
		Expect(err).ToNot(HaveOccurred())
		Expect(label).To(Equal("spec.region"))
		Expect(value).To(Equal("eu"))

		// Default ObjectMeta keys remain accepted.
		_, _, err = scheme.ConvertFieldLabel(gvk, "metadata.name", "a")
		Expect(err).ToNot(HaveOccurred())

		// Unknown keys are rejected.
		_, _, err = scheme.ConvertFieldLabel(gvk, "spec.bogus", "x")
		Expect(err).To(HaveOccurred())
	})

	It("RegisterFieldLabelConversions is a no-op for an empty key set", func() {
		scheme := runtime.NewScheme()
		gvk := schema.GroupVersionKind{Group: "arc", Version: "v1alpha1", Kind: "SelectableObj"}
		Expect(RegisterFieldLabelConversions(scheme, gvk, nil)).To(Succeed())
	})
})
