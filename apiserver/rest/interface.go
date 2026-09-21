// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage/names"
)

type Scoper rest.Scoper

type NameGenerator names.NameGenerator

// AllowCreateOnUpdater implements a subset of rest.RESTUpdateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type AllowCreateOnUpdater interface {
	// AllowCreateOnUpdate returns true if the object can be created by a PUT.
	AllowCreateOnUpdate(ctx context.Context) bool
}

// AllowUnconditionalUpdater implements a subset of rest.RESTUpdateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type AllowUnconditionalUpdater interface {
	// AllowUnconditionalUpdate returns true if the object can be updated
	// unconditionally (irrespective of the latest resource version), when
	// there is no resource version specified in the object.
	AllowUnconditionalUpdate(ctx context.Context) bool
}

// Canonicalizer implements a subset of rest.RESTUpdateStrategy/rest.RESTCreateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type Canonicalizer interface {
	// Canonicalize allows an object to be mutated into a canonical form. This
	// ensures that code that operates on these objects can rely on the common
	// form for things like comparison.  Canonicalize is invoked after
	// validation has succeeded but before the object has been persisted.
	// This method may mutate the object.
	Canonicalize()
}

// PrepareForCreater implements a subset of rest.RESTCreateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type PrepareForCreater interface {
	// PrepareForCreate is invoked on create before validation to normalize
	// the object.  For example: remove fields that are not to be persisted,
	// sort order-insensitive list fields, etc.  This should not remove fields
	// whose presence would be considered a validation error.
	//
	// Often implemented as a type check and an initailization or clearing of
	// status. Clear the status because status changes are internal. External
	// callers of an api (users) should not be setting an initial status on
	// newly created objects.
	PrepareForCreate(ctx context.Context)
}

// PrepareForUpdater implements a subset of rest.RESTUpdateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type PrepareForUpdater interface {
	// PrepareForUpdate is invoked on update before validation to normalize
	// the object.  For example: remove fields that are not to be persisted,
	// sort order-insensitive list fields, etc.  This should not remove fields
	// whose presence would be considered a validation error.
	PrepareForUpdate(ctx context.Context, old runtime.Object)
}

// GenerationTracker opts a resource into Kubernetes generation semantics: DefaultStrategy sets
// metadata.generation to 1 on create and increments it on every update that changes the spec.
//
// This is opt-in rather than automatic because a generic strategy cannot know which part of an
// arbitrary runtime.Object is its "spec", and because turning it on changes what every controller
// watching the resource observes. Upstream Kubernetes does the same work per resource: BeforeCreate
// sets no generation at all, and BeforeUpdate only copies the stored value forward so clients
// cannot forge it, leaving the increment to each resource's own strategy. A resource that does not
// implement this interface therefore stays at generation 0 for its whole life — which silently
// makes every `status.observedGeneration == metadata.generation` comparison a controller performs
// trivially true, so idempotence short-circuits latch and spec edits are never acted on.
//
// Implementations compare their own typed spec, which is exact and needs no reflection:
//
//	func (o *Widget) SpecChanged(old runtime.Object) bool {
//		p, ok := old.(*Widget)
//
//		return !ok || !apiequality.Semantic.DeepEqual(o.Spec, p.Spec)
//	}
//
// Only the spec counts. Metadata-only edits (labels, annotations) must not make controllers
// re-reconcile, and status updates never reach this path: the status subresource installs a
// strategy that calls only its own override.
type GenerationTracker interface {
	// SpecChanged reports whether this object's spec differs from old's. It must report true
	// when old is not of the same type, so an unexpected pairing is treated as a change rather
	// than silently skipping the increment.
	SpecChanged(old runtime.Object) bool
}

// TableConverter implements an adapted version of rest.TableConverter
// it can be used by objects to override DefaultStrategy behaviour.
type TableConverter interface {
	ConvertToTable(ctx context.Context, tableOptions runtime.Object) (*metav1.Table, error)
}

// Validater implements a subset of rest.RESTCreateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type Validater interface {
	// Validate returns an ErrorList with validation errors or nil.  Validate
	// is invoked after default fields in the object have been filled in
	// before the object is persisted.  This method should not mutate the
	// object.
	Validate(ctx context.Context) field.ErrorList
}

// Validater implements a subset of rest.RESTUpdateStrategy and
// it can be used by objects to override DefaultStrategy behaviour.
type ValidateUpdater interface {
	// ValidateUpdate is invoked after default fields in the object have been
	// filled in before the object is persisted.  This method should not mutate
	// the object.
	ValidateUpdate(ctx context.Context, obj runtime.Object) field.ErrorList
}

// ShortNamesProvider allows a resource to specify short names for kubectl.
// Short names allow users to use shorter commands like "kubectl get po" instead of
// "kubectl get pods".
type ShortNamesProvider interface {
	// ShortNames returns a list of short names for the resource.
	ShortNames() []string
}

// SingularNameProvider returns the singular name of the resource.
// This is used by kubectl for discovery and display (e.g., "pod" instead of "pods").
type SingularNameProvider interface {
	// GetSingularName returns the singular form of the resource name.
	GetSingularName() string
}

// SelectableFieldsProvider is an optional interface a resource type may implement
// to contribute additional selectable fields (typically spec fields, e.g.
// "spec.region") for field-selector based list/watch filtering.
//
// The returned set is added to the default ObjectMeta-derived fields
// (metadata.name / metadata.namespace) in GetAttrs. Those ObjectMeta-derived
// keys are reserved: provider fields are additive only and cannot replace them,
// so returning "metadata.name" or "metadata.namespace" here has no effect. Types
// that do not implement this interface behave exactly as before (only ObjectMeta
// fields are selectable).
//
// The keys returned here are advertised to the apiserver's list-options
// conversion automatically, so implementing this interface alone is enough for
// the corresponding field selectors to be accepted. For that derivation to be
// correct the implementation MUST return every selectable key unconditionally
// (with an empty value when unset), matching the upstream Kubernetes convention.
// A resource that needs to advertise a different key set may additionally
// implement SupportedFieldSelectorsProvider, which overrides this default.
type SelectableFieldsProvider interface {
	// SelectableFields returns the object's additional selectable fields.
	SelectableFields() fields.Set
}

// SupportedFieldSelectorsProvider is an optional interface a resource type may
// implement to advertise the field-selector keys it supports beyond the default
// ObjectMeta fields (e.g. "spec.region").
//
// For each advertised key the resource wiring registers a pass-through
// FieldLabelConversionFunc on the scheme for the resource's versioned (and
// internal) GVKs, so the apiserver accepts these selectors during list-options
// conversion instead of rejecting them as unknown.
//
// This is an advanced override: when omitted, the advertised keys default to
// those emitted by SelectableFieldsProvider. Implement it only when the
// advertised key set must differ from the emitted selectable fields.
type SupportedFieldSelectorsProvider interface {
	// SupportedFieldSelectors returns the additional field-selector keys the
	// resource supports (e.g. []string{"spec.region"}).
	SupportedFieldSelectors() []string
}
