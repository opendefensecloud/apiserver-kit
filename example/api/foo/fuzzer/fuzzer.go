package fuzzer

import (
	"go.opendefense.cloud/kit/example/api/foo"
	"sigs.k8s.io/randfill"

	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
)

// Funcs returns the fuzzer functions for the apps api group.
var Funcs = func(codecs runtimeserializer.CodecFactory) []any {
	return []any{
		func(s *foo.BarSpec, c randfill.Continue) {
			c.FillNoCustom(s) // fuzz self without calling this function again
		},
	}
}
