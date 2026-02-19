// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	"go.opendefense.cloud/kit/apiserver/resource"
)

// Storage is an alias for the apiserver's rest.Storage interface.
// It represents a generic storage backend for Kubernetes resources.
type Storage = rest.Storage

// GetAttrs extracts the labels and fields from a runtime.Object for use in storage predicates.
// Returns an error if the object does not implement resource.Object (i.e., lacks metadata).
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	provider, ok := obj.(resource.Object)
	if !ok {
		return nil, nil, fmt.Errorf("given object of type %T does not have metadata", obj)
	}
	om := provider.GetObjectMeta()

	return om.GetLabels(), SelectableFields(om), nil
}

// SelectableFields returns a set of fields (name, namespace, etc.) for the given ObjectMeta.
// Used for field selectors in storage and API queries.
func SelectableFields(obj *metav1.ObjectMeta) fields.Set {
	return generic.ObjectMetaFieldsSet(obj, true)
}

// NewStore constructs a genericregistry.Store for a Kubernetes resource type.
// It wires up the storage strategies, table conversion, and predicate functions.
//
// Parameters:
//   - scheme: runtime.Scheme for type registration
//   - single: function returning a new instance of the resource
//   - list: function returning a new list instance of the resource
//   - gr: GroupResource describing the resource
//   - strategy: Strategy implementation for create/update/delete/table
//   - optsGetter: RESTOptionsGetter for storage backend configuration
//
// Returns:
//   - rest.Storage: configured store for the resource (may be wrapped for ShortNamesProvider)
//   - error: if store setup fails
func NewStore(
	scheme *runtime.Scheme,
	single, list func() runtime.Object,
	gr schema.GroupResource,
	strategy Strategy, optsGetter generic.RESTOptionsGetter) (rest.Storage, error) {
	store := &genericregistry.Store{
		NewFunc:                   single,
		NewListFunc:               list,
		PredicateFunc:             strategy.Match,
		DefaultQualifiedResource:  gr,
		SingularQualifiedResource: gr,
		TableConvertor:            strategy,
		CreateStrategy:            strategy,
		UpdateStrategy:            strategy,
		DeleteStrategy:            strategy,
	}

	// If the strategy implements SingularNameProvider, use the custom singular name.
	if sn, ok := strategy.(SingularNameProvider); ok {
		singularName := sn.GetSingularName()
		if singularName != "" {
			store.SingularQualifiedResource = schema.GroupResource{
				Group:    gr.Group,
				Resource: singularName,
			}
		}
	}

	// If the strategy implements ShortNamesProvider, wrap the store to expose short names.
	if sn, ok := strategy.(ShortNamesProvider); ok && len(sn.ShortNames()) > 0 {
		wrapped := &storeWithShortNames{Store: store, shortNames: sn.ShortNames()}
		options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
		if err := wrapped.CompleteWithOptions(options); err != nil {
			return nil, err
		}

		return wrapped, nil
	}

	// StoreOptions wires up REST options and attribute extraction for filtering.
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return store, nil
}

// storeWithShortNames wraps a genericregistry.Store to provide short names for a resource.
// This implements the ShortNamesProvider interface, allowing kubectl to use short aliases.
type storeWithShortNames struct {
	*genericregistry.Store
	shortNames []string
}

// ShortNames returns the list of short names for the resource.
func (s *storeWithShortNames) ShortNames() []string {
	return s.shortNames
}
