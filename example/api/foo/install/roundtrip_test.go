// Copyright BWI GmbH and apiserver-kit contributors
// SPDX-License-Identifier: Apache-2.0

package install

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/apitesting/roundtrip"

	"go.opendefense.cloud/kit/example/api/foo/fuzzer"
)

func TestRoundTripTypes(t *testing.T) {
	roundtrip.RoundTripTestForAPIGroup(t, Install, fuzzer.Funcs)
	// TODO: enable protobuf generation for the sample-apiserver
	// roundtrip.RoundTripProtobufTestForAPIGroup(t, Install, orderfuzzer.Funcs)
}
