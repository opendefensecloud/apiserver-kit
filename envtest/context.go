// Copyright BWI GmbH and apiserver-kit contributors
// SPDX-License-Identifier: Apache-2.0

package envtest

import (
	"context"

	"github.com/ironcore-dev/ironcore/utils/testing"
)

func Context() context.Context {
	return testing.SetupContext()
}
