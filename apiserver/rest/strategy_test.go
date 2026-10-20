// Copyright BWI GmbH and apiserver-kit contributors
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"context"

	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testObj is a small helper type used to implement several of the
// optional interfaces that DefaultStrategy looks for.
type testObj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Status string
	Flag   bool
}

func (t *testObj) DeepCopyObject() runtime.Object {
	if t == nil {
		return nil
	}
	clone := *t

	return &clone
}

func (t *testObj) GetObjectMeta() *metav1.ObjectMeta { return &t.ObjectMeta }
func (t *testObj) NamespaceScoped() bool             { return true }
func (t *testObj) New() runtime.Object               { return &testObj{} }
func (t *testObj) NewList() runtime.Object           { return &testObjList{} }

func (t *testObj) GetGroupResource() schema.GroupResource {
	return schema.GroupResource{Group: "arc", Resource: "testobjs"}
}

// CopyStatusTo implements resource.ObjectWithStatusSubResource
func (t *testObj) CopyStatusTo(obj runtime.Object) {
	if o, ok := obj.(*testObj); ok {
		o.Status = t.Status
	}
}

// PrepareForCreate implements PrepareForCreater
func (t *testObj) PrepareForCreate(ctx context.Context) { t.Flag = true }

// PrepareForUpdate implements PrepareForUpdater
func (t *testObj) PrepareForUpdate(ctx context.Context, old runtime.Object) { t.Flag = true }

// Validate implements Validater
func (t *testObj) Validate(ctx context.Context) field.ErrorList {
	return field.ErrorList{field.Invalid(field.NewPath("spec"), "bad", "invalid")}
}

// ValidateUpdate implements ValidateUpdater
func (t *testObj) ValidateUpdate(ctx context.Context, old runtime.Object) field.ErrorList {
	return field.ErrorList{field.Invalid(field.NewPath("spec"), "bad", "invalid")}
}

// Canonicalize implements Canonicalizer
func (t *testObj) Canonicalize() { t.Flag = true }

// ConvertToTable implements TableConverter
func (t *testObj) ConvertToTable(ctx context.Context, _ runtime.Object) (*metav1.Table, error) {
	return &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string"},
			{Name: "Status", Type: "string"},
		},
		Rows: []metav1.TableRow{
			{
				Cells: []any{t.Name, t.Status},
			},
		},
	}, nil
}

// testObjList is a minimal list type returned by NewList.
type testObjList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []testObj
}

func (t *testObjList) DeepCopyObject() runtime.Object {
	if t == nil {
		return nil
	}
	clone := *t

	return &clone
}

// testObjListWithConvertor is a list type that implements ConvertToTable
// with different columns than testObj.
type testObjListWithConvertor struct {
	testObjList
}

// ConvertToTable implements TableConverter for testObjListWithConvertor
func (t *testObjListWithConvertor) ConvertToTable(ctx context.Context, _ runtime.Object) (*metav1.Table, error) {
	return &metav1.Table{
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "Count", Type: "integer"},
			{Name: "Resource", Type: "string"},
		},
		Rows: []metav1.TableRow{
			{
				Cells: []any{len(t.Items), "testobjs"},
			},
		},
	}, nil
}

// nameGen implements NameGenerator
type nameGen struct {
	testObj
}

func (n *nameGen) GenerateName(base string) string { return base + "-GEN" }

// scoper implements Scoper
type scoper struct {
	testObj
}

func (s *scoper) NamespaceScoped() bool { return false }

// allowCreate implements AllowCreateOnUpdater
type allowCreate struct {
	testObj
}

func (a *allowCreate) AllowCreateOnUpdate(_ context.Context) bool { return true }

// allowUnconditional implements AllowUnconditionalUpdater
type allowUnconditional struct {
	testObj
}

func (a *allowUnconditional) AllowUnconditionalUpdate(_ context.Context) bool { return true }

// plainObj is a minimal runtime.Object carrying only ObjectMeta, used to test
// default name validation in isolation. It intentionally implements none of
// the optional strategy interfaces.
type plainObj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (p *plainObj) DeepCopyObject() runtime.Object {
	if p == nil {
		return nil
	}
	clone := *p

	return &clone
}

// labelNameObj implements NameValidator with DNS-1123 label rules.
type labelNameObj struct {
	plainObj
}

func (l *labelNameObj) ValidateName(name string, prefix bool) []string {
	return apimachineryvalidation.NameIsDNSLabel(name, prefix)
}

var _ = Describe("DefaultStrategy", func() {
	It("should use NameGenerator for GenerateName", func() {
		ds := DefaultStrategy{Object: &nameGen{}}
		Expect(ds.GenerateName("base")).To(Equal("base-GEN"))
	})

	It("should use Scoper for NamespaceScoped", func() {
		ds := DefaultStrategy{Object: &scoper{}}
		Expect(ds.NamespaceScoped()).To(BeFalse())
	})

	It("should call PrepareForCreater on PrepareForCreate", func() {
		obj := &testObj{}
		ds := DefaultStrategy{}
		ds.PrepareForCreate(context.Background(), obj)
		Expect(obj.Flag).To(BeTrue())
	})

	It("should copy status and call PrepareForUpdater on PrepareForUpdate", func() {
		old := &testObj{Status: "old-status"}
		obj := &testObj{Status: "new-status"}
		ds := DefaultStrategy{}
		ds.PrepareForUpdate(context.Background(), obj, old)
		Expect(obj.Status).To(Equal("old-status"))
		Expect(obj.Flag).To(BeTrue())
	})

	It("should delegate Validate and ValidateUpdate to object", func() {
		obj := &testObj{}
		ds := DefaultStrategy{}
		Expect(ds.Validate(context.Background(), obj)).ToNot(BeEmpty())
		Expect(ds.ValidateUpdate(context.Background(), obj, &testObj{})).ToNot(BeEmpty())
	})

	It("should delegate AllowCreateOnUpdate and AllowUnconditionalUpdate", func() {
		ds1 := DefaultStrategy{Object: &allowCreate{}}
		Expect(ds1.AllowCreateOnUpdate(context.Background())).To(BeTrue())
		ds2 := DefaultStrategy{Object: &allowUnconditional{}}
		Expect(ds2.AllowUnconditionalUpdate(context.Background())).To(BeTrue())
	})

	It("should delegate Canonicalize and ConvertToTable to object", func() {
		obj := &testObj{Name: "test-obj", Status: "ready"}
		ds := DefaultStrategy{Object: obj}
		ds.Canonicalize(obj)
		Expect(obj.Flag).To(BeTrue())
		tbl, err := ds.ConvertToTable(context.Background(), obj, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(tbl).ToNot(BeNil())
		Expect(tbl.ColumnDefinitions).To(HaveLen(2))
		Expect(tbl.ColumnDefinitions[0].Name).To(Equal("Name"))
		Expect(tbl.ColumnDefinitions[1].Name).To(Equal("Status"))
		Expect(tbl.Rows).To(HaveLen(1))
		Expect(tbl.Rows[0].Cells).To(Equal([]any{"test-obj", "ready"}))
	})

	It("should use testObj's ConvertToTable implementation with DefaultStrategy", func() {
		obj := &testObj{Name: "my-object", Status: "active"}
		ds := NewDefaultStrategy(obj, nil, schema.GroupResource{Group: "arc", Resource: "testobjs"})
		tbl, err := ds.ConvertToTable(context.Background(), obj, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(tbl).ToNot(BeNil())
		Expect(tbl.ColumnDefinitions).To(HaveLen(2))
		Expect(tbl.ColumnDefinitions[0].Name).To(Equal("Name"))
		Expect(tbl.ColumnDefinitions[1].Name).To(Equal("Status"))
		Expect(tbl.Rows).To(HaveLen(1))
		Expect(tbl.Rows[0].Cells).To(Equal([]any{"my-object", "active"}))
	})

	It("should use testObj's ConvertToTable for items in testObjList", func() {
		list := &testObjList{
			Items: []testObj{
				{Name: "obj1", Status: "ready"},
				{Name: "obj2", Status: "pending"},
			},
		}
		ds := NewDefaultStrategy(&testObj{}, nil, schema.GroupResource{Group: "arc", Resource: "testobjs"})
		tbl, err := ds.ConvertToTable(context.Background(), list, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(tbl).ToNot(BeNil())
		// Each item should contribute one row, so total 2 rows
		Expect(tbl.Rows).To(HaveLen(2))
		// Verify column definitions are present
		Expect(tbl.ColumnDefinitions).To(HaveLen(2))
		Expect(tbl.ColumnDefinitions[0].Name).To(Equal("Name"))
		Expect(tbl.ColumnDefinitions[1].Name).To(Equal("Status"))
		// Verify row data
		Expect(tbl.Rows[0].Cells).To(Equal([]any{"obj1", "ready"}))
		Expect(tbl.Rows[1].Cells).To(Equal([]any{"obj2", "pending"}))
	})

	It("should use list's ConvertToTable implementation if explicitly implemented", func() {
		list := &testObjListWithConvertor{
			Items: []testObj{
				{Name: "obj1", Status: "ready"},
				{Name: "obj2", Status: "pending"},
				{Name: "obj3", Status: "running"},
			},
		}
		ds := NewDefaultStrategy(&testObj{}, nil, schema.GroupResource{Group: "arc", Resource: "testobjs"})
		tbl, err := ds.ConvertToTable(context.Background(), list, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(tbl).ToNot(BeNil())
		// Should use list's ConvertToTable, which returns a single row
		Expect(tbl.Rows).To(HaveLen(1))
		// Verify column definitions are different from testObj's columns
		Expect(tbl.ColumnDefinitions).To(HaveLen(2))
		Expect(tbl.ColumnDefinitions[0].Name).To(Equal("Count"))
		Expect(tbl.ColumnDefinitions[1].Name).To(Equal("Resource"))
		// Verify row data shows count and resource type
		Expect(tbl.Rows[0].Cells).To(Equal([]any{3, "testobjs"}))
	})
})

var _ = Describe("PrepareForUpdaterStrategy", func() {
	It("should call OverrideFn on PrepareForUpdate", func() {
		called := false
		var gotCtx context.Context
		var gotObj, gotOld runtime.Object
		s := &PrepareForUpdaterStrategy{
			RESTUpdateStrategy: &DefaultStrategy{Object: &testObj{}},
			OverrideFn: func(ctx context.Context, obj, old runtime.Object) {
				called = true
				gotCtx = ctx
				gotObj = obj
				gotOld = old
			},
		}
		obj := &testObj{Status: "new"}
		old := &testObj{Status: "old"}
		//nolint:staticcheck
		ctx := context.WithValue(context.Background(), "key", "val")
		s.PrepareForUpdate(ctx, obj, old)
		Expect(called).To(BeTrue())
		Expect(gotCtx).To(Equal(ctx))
		Expect(gotObj).To(Equal(obj))
		Expect(gotOld).To(Equal(old))
	})

	It("should not panic if OverrideFn is nil", func() {
		s := &PrepareForUpdaterStrategy{}
		obj := &testObj{}
		old := &testObj{}
		// Should not panic, but does nothing
		Expect(func() { s.PrepareForUpdate(context.Background(), obj, old) }).ToNot(Panic())
	})
})

// genObj opts into generation semantics. It deliberately does NOT implement
// resource.ObjectWithStatusSubResource, so these specs exercise the generation path alone.
type genObj struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec string
}

func (g *genObj) DeepCopyObject() runtime.Object {
	if g == nil {
		return nil
	}
	clone := *g

	return &clone
}

func (g *genObj) GetObjectMeta() *metav1.ObjectMeta { return &g.ObjectMeta }

// SpecChanged implements GenerationTracker.
func (g *genObj) SpecChanged(old runtime.Object) bool {
	o, ok := old.(*genObj)

	return !ok || g.Spec != o.Spec
}

var _ = Describe("DefaultStrategy generation semantics", func() {
	// Upstream BeforeCreate never sets a generation and BeforeUpdate only copies the stored one
	// forward, so incrementing it is each strategy's job. Without this a resource sits at
	// generation 0 forever and every status.observedGeneration == metadata.generation test a
	// controller makes is trivially true.
	It("starts an opted-in resource at generation 1 on create", func() {
		obj := &genObj{Spec: "a"}
		DefaultStrategy{}.PrepareForCreate(context.Background(), obj)
		Expect(obj.GetGeneration()).To(Equal(int64(1)))
	})

	It("leaves a resource that has not opted in at generation 0", func() {
		obj := &testObj{}
		DefaultStrategy{}.PrepareForCreate(context.Background(), obj)
		Expect(obj.GetGeneration()).To(Equal(int64(0)))
	})

	It("increments the generation when the spec changes", func() {
		old := &genObj{Spec: "a"}
		old.SetGeneration(7)
		// BeforeUpdate has already copied the stored generation onto the incoming object.
		obj := &genObj{Spec: "b"}
		obj.SetGeneration(7)
		DefaultStrategy{}.PrepareForUpdate(context.Background(), obj, old)
		Expect(obj.GetGeneration()).To(Equal(int64(8)))
	})

	// A label or annotation edit must not make every controller re-reconcile.
	It("leaves the generation alone when the spec is unchanged", func() {
		old := &genObj{Spec: "a"}
		old.SetGeneration(7)
		obj := &genObj{Spec: "a"}
		obj.SetGeneration(7)
		obj.SetLabels(map[string]string{"k": "v"})
		DefaultStrategy{}.PrepareForUpdate(context.Background(), obj, old)
		Expect(obj.GetGeneration()).To(Equal(int64(7)))
	})

	It("leaves a resource that has not opted in alone on update", func() {
		obj := &testObj{Status: "new"}
		old := &testObj{Status: "old"}
		old.SetGeneration(3)
		obj.SetGeneration(3)
		DefaultStrategy{}.PrepareForUpdate(context.Background(), obj, old)
		Expect(obj.GetGeneration()).To(Equal(int64(3)))
	})
})

var _ = Describe("DefaultStrategy name validation", func() {
	ds := DefaultStrategy{}
	ctx := context.Background()

	findErr := func(errs field.ErrorList, fieldPath string) (field.Error, bool) {
		for _, e := range errs {
			if e.Field == fieldPath {
				return *e, true
			}
		}

		return field.Error{}, false
	}

	expectNameError := func(errs field.ErrorList) {
		e, found := findErr(errs, "metadata.name")
		Expect(found).To(BeTrue(), "expected a metadata.name error")
		Expect(e.Type).To(Equal(field.ErrorTypeInvalid))
	}

	It("should accept valid DNS-1123 subdomain names by default", func() {
		for _, name := range []string{"valid-name", "valid.name", "a", "valid-name-1"} {
			obj := &plainObj{Name: name}
			Expect(ds.Validate(ctx, obj)).To(BeEmpty())
		}
	})

	It("should reject names with spaces, underscores, or uppercase characters by default", func() {
		for _, name := range []string{"has space", "under_score", "Upper-Case"} {
			obj := &plainObj{Name: name}
			expectNameError(ds.Validate(ctx, obj))
		}
	})

	It("should require a name or generateName when the name is empty", func() {
		obj := &plainObj{}
		errs := ds.Validate(ctx, obj)
		e, found := findErr(errs, "metadata.name")
		Expect(found).To(BeTrue(), "expected a metadata.name error")
		Expect(e.Type).To(Equal(field.ErrorTypeRequired))
		Expect(e.Detail).To(Equal("name or generateName is required"))
	})

	It("should not require a name when a generateName is given", func() {
		// The store expands generateName into name before validation.
		obj := &plainObj{Name: "valid-prefix-x7k2q"}
		obj.GenerateName = "valid-prefix-"
		Expect(ds.Validate(ctx, obj)).To(BeEmpty())
	})

	It("should validate generateName as a prefix", func() {
		// A trailing dash is valid for a generateName prefix.
		obj := &plainObj{Name: "foo-x7k2q"}
		obj.GenerateName = "foo-"
		Expect(ds.Validate(ctx, obj)).To(BeEmpty())

		// With label rules, a dotted prefix is invalid and reported on
		// metadata.generateName, in addition to the generated name.
		lbl := &labelNameObj{Name: "my-app-x7k2q"}
		lbl.GenerateName = "my.app-"
		errs := ds.Validate(ctx, lbl)
		_, found := findErr(errs, "metadata.generateName")
		Expect(found).To(BeTrue(), "expected a metadata.generateName error")
	})

	It("should use the object's NameValidator rules instead of DNS-1123 subdomain", func() {
		// Valid subdomain but not a valid label.
		obj := &labelNameObj{Name: "has.dots"}
		expectNameError(ds.Validate(ctx, obj))

		// Valid label, also a valid subdomain.
		obj = &labelNameObj{Name: "valid-label"}
		Expect(ds.Validate(ctx, obj)).To(BeEmpty())
	})

	It("should reject a trailing dash in a final name under label rules", func() {
		// The same string that is a valid prefix is not a valid final name.
		obj := &labelNameObj{Name: "foo-"}
		expectNameError(ds.Validate(ctx, obj))
	})

	It("should not validate the name on update, only delegate to ValidateUpdater", func() {
		// Names are validated on create only; updates must not be blocked.
		plain := &plainObj{Name: "has space"}
		Expect(ds.ValidateUpdate(ctx, plain, &plainObj{Name: "has space"})).To(BeEmpty())

		// Only the object's own ValidateUpdater errors are returned.
		obj := &testObj{Name: "bad_name"}
		errs := ds.ValidateUpdate(ctx, obj, &testObj{})
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Field).To(Equal("spec"))
	})

	It("should aggregate name errors with Validater errors on create", func() {
		obj := &testObj{Name: "bad_name"}
		errs := ds.Validate(ctx, obj)
		Expect(errs).To(HaveLen(2))
		Expect(errs[0].Field).To(Equal("metadata.name"))
		Expect(errs[1].Field).To(Equal("spec"))
	})
})
