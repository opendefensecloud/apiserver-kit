// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/server"

	"go.opendefense.cloud/kit/apiserver/resource"
	"go.opendefense.cloud/kit/apiserver/rest"
)

// ResourceHandler holds the configuration for registering a resource with the API server.
type ResourceHandler struct {
	groupVersions []schema.GroupVersion
	apiGroupFn    APIGroupFn
}

// Resource registers a Kubernetes resource with the API server.
//
// The type parameters are:
//   - E: the internal resource type implementing resource.Object
//   - T: the typed resource (e.g., *Bar) that also implements resource.ObjectWithDeepCopy[E]
//
// The gvs parameter specifies which group versions to register.
//
// To customize the resource's short names or singular name in kubectl, implement
// ShortNamesProvider or SingularNameProvider on the resource type T:
//
//	func (b *Bar) ShortNames() []string {
//	    return []string{"br"}
//	}
//
//	func (b *Bar) GetSingularName() string {
//	    return "bar"
//	}
func Resource[E resource.Object, T resource.ObjectWithDeepCopy[E]](obj T, gvs ...schema.GroupVersion) ResourceHandler {
	return ResourceHandler{
		groupVersions: gvs,
		apiGroupFn: func(scheme *runtime.Scheme, codecs serializer.CodecFactory, c *server.CompletedConfig) server.APIGroupInfo {
			gr := obj.GetGroupResource()
			strategy := rest.NewDefaultStrategy(obj, scheme, gr)
			store, err := rest.NewStore(scheme, obj.New, obj.NewList, gr, strategy, c.RESTOptionsGetter)
			if err != nil {
				panic(err)
			}

			storage := map[string]rest.Storage{}
			storage[gr.Resource] = store

			if _, ok := any(obj).(resource.ObjectWithStatusSubResource); ok {
				statusPrepareForUpdate := func(ctx context.Context, obj, old runtime.Object) {
					// We copy status to old
					statusObj := any(obj).(resource.ObjectWithStatusSubResource)
					statusObj.CopyStatusTo(old)
					// And use old (with new status) to reset spec of new obj
					copyableObj := any(obj).(E)
					copyableOld := any(old).(T)
					copyableOld.DeepCopyInto(copyableObj)
				}
				// We need to access the underlying *registry.Store for status subresource.
				// Use rest.Unwrap to handle both wrapped (storeWithShortNames) and unwrapped cases.
				// Make a value copy so we can modify only the status copy's UpdateStrategy.
				statusStore := *rest.Unwrap(store)
				statusStore.UpdateStrategy = &rest.PrepareForUpdaterStrategy{
					RESTUpdateStrategy: statusStore.UpdateStrategy,
					OverrideFn:         statusPrepareForUpdate,
				}
				storage[gr.Resource+"/status"] = &statusStore
			}

			apiGroupInfo := server.NewDefaultAPIGroupInfo(gr.Group, scheme, metav1.ParameterCodec, codecs)

			for _, gv := range gvs {
				if gv.Group != gr.Group {
					panic("unexpected group mismatch")
				}
				apiGroupInfo.VersionedResourcesStorageMap[gv.Version] = storage
			}

			return apiGroupInfo
		},
	}
}
