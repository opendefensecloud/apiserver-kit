// Copyright 2026 BWI GmbH and contributors
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"go.opendefense.cloud/kit/envtest"
	"go.opendefense.cloud/kit/example/api/foo/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bar", func() {
	var (
		ctx = envtest.Context()
		ns  = SetupTest(ctx)
		bar = &v1alpha1.Bar{}
	)

	Context("Bar", func() {
		It("should allow creating a bar", func() {
			By("creating a test bar")
			bar = &v1alpha1.Bar{
				Namespace:    ns.Name,
				GenerateName: "test-",
				Spec:         v1alpha1.BarSpec{},
			}
			Expect(k8sClient.Create(ctx, bar)).To(Succeed())
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(bar), bar)).To(Succeed())
		})
		It("should allow deleting an bar", func() {
			By("deleting a test bar")
			Expect(k8sClient.Delete(ctx, bar)).To(Succeed())
		})
	})

})

var _ = Describe("Bar", func() {
	var (
		ctx = envtest.Context()
		bar = &v1alpha1.ClusterBar{}
	)
	Context("ClusterBar", func() {
		It("should allow creating a bar", func() {
			By("creating a test bar")
			bar = &v1alpha1.ClusterBar{
				GenerateName: "test-",
				Spec:         v1alpha1.BarSpec{},
			}
			Expect(k8sClient.Create(ctx, bar)).To(Succeed())
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(bar), bar)).To(Succeed())
		})
		It("should allow deleting an bar", func() {
			By("deleting a test bar")
			Expect(k8sClient.Delete(ctx, bar)).To(Succeed())
		})
	})
})
