// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"fmt"
	"maps"
	"slices"
	"strings"

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
//
// The returned field set always contains the default ObjectMeta-derived fields
// (metadata.name / metadata.namespace). If the object additionally implements
// SelectableFieldsProvider, its contributed fields (typically spec fields) are
// merged on top, enabling spec field-selector filtering. Objects that do not
// implement the interface behave exactly as before.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	provider, ok := obj.(resource.Object)
	if !ok {
		return nil, nil, fmt.Errorf("given object of type %T does not have metadata", obj)
	}
	om := provider.GetObjectMeta()

	fieldSet := SelectableFields(om)
	if sfp, ok := obj.(SelectableFieldsProvider); ok {
		maps.Copy(fieldSet, sfp.SelectableFields())
	}

	return om.GetLabels(), fieldSet, nil
}

// SelectableFields returns a set of fields (name, namespace, etc.) for the given ObjectMeta.
// Used for field selectors in storage and API queries.
func SelectableFields(obj *metav1.ObjectMeta) fields.Set {
	return generic.ObjectMetaFieldsSet(obj, true)
}

// FieldSelectorKeys returns the additional field-selector keys a resource
// advertises beyond the default ObjectMeta fields, applying the following
// precedence:
//
//   - If obj implements SupportedFieldSelectorsProvider, its explicit key set is
//     used (advanced override — e.g. a key set that differs from the emitted
//     selectable fields).
//   - Otherwise, if obj implements SelectableFieldsProvider, the keys are derived
//     from the emitted fields (sorted for deterministic registration). This
//     assumes SelectableFields emits every selectable key unconditionally, per
//     the upstream Kubernetes convention.
//   - Otherwise nil is returned and the resource keeps default behavior.
func FieldSelectorKeys(obj any) []string {
	if fsp, ok := obj.(SupportedFieldSelectorsProvider); ok {
		return fsp.SupportedFieldSelectors()
	}
	if sfp, ok := obj.(SelectableFieldsProvider); ok {
		return slices.Sorted(maps.Keys(sfp.SelectableFields()))
	}

	return nil
}

// RegisterFieldLabelConversions registers pass-through FieldLabelConversionFuncs on
// the scheme for the given GVKs and field-selector keys.
//
// A field selector supplied on a list/watch request is validated against the
// scheme during list-options conversion; any key not registered via
// AddFieldLabelConversionFunc is rejected as unknown. This helper registers a
// pass-through conversion (identity: the label==value, the key unchanged) for
// every supported key, plus the always-supported ObjectMeta keys
// (metadata.name / metadata.namespace), so the apiserver accepts them.
//
// It is a no-op when keys is empty, keeping default behavior unchanged for
// resources that do not advertise extra selectors.
func RegisterFieldLabelConversions(scheme *runtime.Scheme, gvk schema.GroupVersionKind, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	supported := map[string]struct{}{
		"metadata.name":      {},
		"metadata.namespace": {},
	}
	for _, k := range keys {
		supported[k] = struct{}{}
	}

	return scheme.AddFieldLabelConversionFunc(gvk, func(label, value string) (string, string, error) {
		if _, ok := supported[label]; ok {
			return label, value, nil
		}

		return "", "", fmt.Errorf("field label not supported: %s", label)
	})
}

// GroupScopedResourcePrefix returns the etcd storage ResourcePrefix for a
// resource, scoped by its API group as "<group>/<resource>" (resource
// lowercased). Scoping every object's key by its group keeps the storage layout
// independent of how many groups the server exposes: two groups may share a
// resource name without colliding, and the key does not move when a server goes
// from exposing one group to several.
func GroupScopedResourcePrefix(gr schema.GroupResource) string {
	return gr.Group + "/" + strings.ToLower(gr.Resource)
}

// groupScopedRESTOptionsGetter wraps a RESTOptionsGetter to force each resource's
// storage ResourcePrefix to GroupScopedResourcePrefix, so the group is encoded in
// the object's key rather than relying on a group-specific storage root.
type groupScopedRESTOptionsGetter struct {
	delegate       generic.RESTOptionsGetter
	resourcePrefix string
}

func (g groupScopedRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (generic.RESTOptions, error) {
	opts, err := g.delegate.GetRESTOptions(resource, example)
	if err != nil {
		return opts, err
	}
	opts.ResourcePrefix = g.resourcePrefix

	return opts, nil
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
	// Scope the storage key by API group so multiple groups can be served from
	// one etcd root without resource-name collisions.
	optsGetter = groupScopedRESTOptionsGetter{delegate: optsGetter, resourcePrefix: GroupScopedResourcePrefix(gr)}
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

// Unwrap returns the underlying *genericregistry.Store.
// This is useful when you need to access the store directly, e.g., for setting
// the status subresource update strategy.
func Unwrap(s rest.Storage) *genericregistry.Store {
	if wrapped, ok := s.(*storeWithShortNames); ok {
		return wrapped.Store
	}

	return s.(*genericregistry.Store)
}
