// Copyright 2026 BWI GmbH and contributors
// SPDX-License-Identifier: Apache-2.0

package fuzzer

import (
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/randfill"

	"go.opendefense.cloud/kit/example/api/foo"
)

// Funcs returns the fuzzer functions for the apps api group.
var Funcs = func(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(s *foo.BarSpec, c randfill.Continue) {
			c.FillNoCustom(s) // fuzz self without calling this function again
		},
	}
}
