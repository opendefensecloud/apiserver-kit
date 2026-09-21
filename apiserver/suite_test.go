// Copyright BWI GmbH and apiserver-kit contributors
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "APIServer Suite")
}
