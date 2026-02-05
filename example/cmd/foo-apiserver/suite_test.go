// Copyright 2026 BWI GmbH and contributors
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"go.opendefense.cloud/kit/envtest"
	"go.opendefense.cloud/kit/example/api/foo/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	pollingInterval      = 50 * time.Millisecond
	eventuallyTimeout    = 3 * time.Second
	consistentlyDuration = 1 * time.Second
	apiServiceTimeout    = 5 * time.Minute
)

var (
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func TestAPIServer(t *testing.T) {
	SetDefaultConsistentlyPollingInterval(pollingInterval)
	SetDefaultEventuallyPollingInterval(pollingInterval)
	SetDefaultEventuallyTimeout(eventuallyTimeout)
	SetDefaultConsistentlyDuration(consistentlyDuration)

	RegisterFailHandler(Fail)

	RunSpecs(t, "Foo API Server Suite")
}

var _ = BeforeSuite(func() {
	var err error

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")

	Expect(v1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())

	testEnv, err = envtest.NewEnvironment(
		"go.opendefense.cloud/kit/example/cmd/foo-apiserver",
		[]string{},
		[]string{filepath.Join("..", "..", "test", "fixtures")},
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(testEnv).NotTo(BeNil())

	k8sClient, err = testEnv.Start(scheme.Scheme, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(testEnv.Stop)

	Expect(testEnv.WaitUntilReadyWithTimeout(apiServiceTimeout)).To(Succeed())
})

func SetupTest(ctx context.Context) *corev1.Namespace {
	var (
		ns = &corev1.Namespace{}
	)

	BeforeEach(func() {
		*ns = corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "testns-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed(), "failed to create test namespace")
		DeferCleanup(k8sClient.Delete, ctx, ns)
	})

	return ns
}
