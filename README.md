# apiserver-kit

A Go library for building Kubernetes-style aggregated API servers with minimal boilerplate.

## Overview

This library provides:

- **Builder pattern** for constructing API servers with etcd storage, TLS, admission, and OpenAPI support
- **Generic resource registration** that automatically configures storage and status subresources
- **Strategy interfaces** for customizing validation, normalization, and update semantics
- **Test environment** wrapper for integration testing with aggregated API servers

## Installation

```bash
go get go.opendefense.cloud/kit
```

## Quick Start

### 1. Define your resource type

Your resource must implement `resource.Object`:

```go
package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
)

type MyResource struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec              MyResourceSpec   `json:"spec,omitempty"`
    Status            MyResourceStatus `json:"status,omitempty"`
}

// Required: resource.Object interface
func (m *MyResource) GetObjectMeta() *metav1.ObjectMeta { return &m.ObjectMeta }
func (m *MyResource) NamespaceScoped() bool             { return true }
func (m *MyResource) New() runtime.Object               { return &MyResource{} }
func (m *MyResource) NewList() runtime.Object           { return &MyResourceList{} }
func (m *MyResource) GetGroupResource() schema.GroupResource {
    return schema.GroupResource{Group: "mygroup.example.com", Resource: "myresources"}
}

// Optional: enable /status subresource
func (m *MyResource) CopyStatusTo(obj runtime.Object) {
    obj.(*MyResource).Status = m.Status
}

// Optional: set singularName for kubectl usage
func (m *MyResource) GetSingularName() string {
    return "myresource"
}

// Optional: set shortNames for kubectl usage
func (m *MyResource) ShortNames() []string {
    return []string{"mr", "mrs"}
}
```

### 2. Build and run the API server

```go
package main

import (
    "os"

    "go.opendefense.cloud/kit/apiserver"
    "k8s.io/apimachinery/pkg/runtime"

    myv1alpha1 "example.com/myproject/api/v1alpha1"
)

func main() {
    scheme := runtime.NewScheme()
    myv1alpha1.AddToScheme(scheme)

    os.Exit(apiserver.NewBuilder(scheme).
        WithComponentName("myapi").
        WithGroupVersions(myv1alpha1.SchemeGroupVersion).
        With(apiserver.Resource[*myv1alpha1.MyResource](&myv1alpha1.MyResource{}, myv1alpha1.SchemeGroupVersion)).
        WithOpenAPIDefinitions("My API", "v1alpha1", myv1alpha1.GetOpenAPIDefinitions).
        Execute())
}
```

### 3. Integration testing with envtest

```go
package myresource_test

import (
    "testing"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"

    "go.opendefense.cloud/kit/envtest"
)

var testEnv *envtest.Environment

func TestMyResource(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "MyResource Suite")
}

var _ = BeforeSuite(func() {
    var err error
    testEnv, err = envtest.NewEnvironment(
        "path/to/cmd/apiserver",           // main.go path
        []string{"path/to/crds"},          // CRD directories
        []string{"path/to/apiservices"},   // APIService directories
    )
    Expect(err).NotTo(HaveOccurred())

    _, err = testEnv.Start(scheme, GinkgoWriter)
    Expect(err).NotTo(HaveOccurred())

    err = testEnv.WaitUntilReadyWithTimeout(30 * time.Second)
    Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
    Expect(testEnv.Stop()).To(Succeed())
})
```

## Customizing Resource Behavior

Resources can implement optional interfaces to customize API server behavior:

| Interface                   | Purpose                               |
| ---                         | ---                                   |
| `Validater`                 | Validate on create                    |
| `ValidateUpdater`           | Validate on update                    |
| `PrepareForCreater`         | Normalize before create               |
| `PrepareForUpdater`         | Normalize before update               |
| `Canonicalizer`             | Transform to canonical form           |
| `AllowCreateOnUpdater`      | Allow PUT to create                   |
| `AllowUnconditionalUpdater` | Allow updates without resourceVersion |
| `TableConverter`            | Custom kubectl table output           |
| `ShortNamesProvider`        | Custom short names for the resource   |
| `SingularNameProvider`      | Define the singular name              |

Example validation:

```go
func (m *MyResource) Validate(ctx context.Context) field.ErrorList {
    var errs field.ErrorList
    if m.Spec.Name == "" {
        errs = append(errs, field.Required(field.NewPath("spec", "name"), "name is required"))
    }
    return errs
}
```

## Project Structure

```
apiserver/
├── builder.go       # Builder pattern for API server construction
├── resource.go      # Generic Resource() function for registration
├── resource/
│   └── object.go    # Core Object interface definitions
└── rest/
    ├── rest.go      # Storage creation utilities
    ├── strategy.go  # DefaultStrategy implementation
    └── interface.go # Optional behavior interfaces

envtest/
├── environment.go   # Test environment wrapper
└── context.go       # Test context utilities
```

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Format code
make fmt

# Update dependencies
make mod
```

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.
