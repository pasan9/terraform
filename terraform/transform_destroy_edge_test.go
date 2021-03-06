package terraform

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/states"
)

func TestDestroyEdgeTransformer_basic(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}
	g.Add(testDestroyNode("test_object.A"))
	g.Add(testDestroyNode("test_object.B"))

	state := states.NewState()
	root := state.EnsureModule(addrs.RootModuleInstance)
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.A").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"A"}`),
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.B").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:       states.ObjectReady,
			AttrsJSON:    []byte(`{"id":"B","test_string":"x"}`),
			Dependencies: []addrs.ConfigResource{mustResourceAddr("test_object.A")},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	if err := (&AttachStateTransformer{State: state}).Transform(&g); err != nil {
		t.Fatal(err)
	}

	tf := &DestroyEdgeTransformer{
		Config:  testModule(t, "transform-destroy-edge-basic"),
		Schemas: simpleTestSchemas(),
	}
	if err := tf.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformDestroyEdgeBasicStr)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

func TestDestroyEdgeTransformer_multi(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}
	g.Add(testDestroyNode("test_object.A"))
	g.Add(testDestroyNode("test_object.B"))
	g.Add(testDestroyNode("test_object.C"))

	state := states.NewState()
	root := state.EnsureModule(addrs.RootModuleInstance)
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.A").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"A"}`),
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.B").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:       states.ObjectReady,
			AttrsJSON:    []byte(`{"id":"B","test_string":"x"}`),
			Dependencies: []addrs.ConfigResource{mustResourceAddr("test_object.A")},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.C").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"C","test_string":"x"}`),
			Dependencies: []addrs.ConfigResource{
				mustResourceAddr("test_object.A"),
				mustResourceAddr("test_object.B"),
			},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)

	if err := (&AttachStateTransformer{State: state}).Transform(&g); err != nil {
		t.Fatal(err)
	}

	tf := &DestroyEdgeTransformer{
		Config:  testModule(t, "transform-destroy-edge-multi"),
		Schemas: simpleTestSchemas(),
	}
	if err := tf.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformDestroyEdgeMultiStr)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

func TestDestroyEdgeTransformer_selfRef(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}
	g.Add(testDestroyNode("test_object.A"))
	tf := &DestroyEdgeTransformer{
		Config:  testModule(t, "transform-destroy-edge-self-ref"),
		Schemas: simpleTestSchemas(),
	}
	if err := tf.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformDestroyEdgeSelfRefStr)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

func TestDestroyEdgeTransformer_module(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}
	g.Add(testDestroyNode("module.child.test_object.b"))
	g.Add(testDestroyNode("test_object.a"))
	state := states.NewState()
	root := state.EnsureModule(addrs.RootModuleInstance)
	child := state.EnsureModule(addrs.RootModuleInstance.Child("child", addrs.NoKey))
	root.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.a").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:       states.ObjectReady,
			AttrsJSON:    []byte(`{"id":"a"}`),
			Dependencies: []addrs.ConfigResource{mustResourceAddr("module.child.test_object.b")},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	child.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.b").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"b","test_string":"x"}`),
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)

	if err := (&AttachStateTransformer{State: state}).Transform(&g); err != nil {
		t.Fatal(err)
	}

	tf := &DestroyEdgeTransformer{
		Config:  testModule(t, "transform-destroy-edge-module"),
		Schemas: simpleTestSchemas(),
	}
	if err := tf.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testTransformDestroyEdgeModuleStr)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

func TestDestroyEdgeTransformer_moduleOnly(t *testing.T) {
	g := Graph{Path: addrs.RootModuleInstance}
	g.Add(testDestroyNode("module.child.test_object.a"))
	g.Add(testDestroyNode("module.child.test_object.b"))
	g.Add(testDestroyNode("module.child.test_object.c"))

	state := states.NewState()
	child := state.EnsureModule(addrs.RootModuleInstance.Child("child", addrs.NoKey))
	child.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.a").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"a"}`),
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	child.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.b").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:       states.ObjectReady,
			AttrsJSON:    []byte(`{"id":"b","test_string":"x"}`),
			Dependencies: []addrs.ConfigResource{mustResourceAddr("module.child.test_object.a")},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)
	child.SetResourceInstanceCurrent(
		mustResourceInstanceAddr("test_object.c").Resource,
		&states.ResourceInstanceObjectSrc{
			Status:    states.ObjectReady,
			AttrsJSON: []byte(`{"id":"c","test_string":"x"}`),
			Dependencies: []addrs.ConfigResource{
				mustResourceAddr("module.child.test_object.a"),
				mustResourceAddr("module.child.test_object.b"),
			},
		},
		mustProviderConfig(`provider["registry.terraform.io/hashicorp/test"]`),
	)

	if err := (&AttachStateTransformer{State: state}).Transform(&g); err != nil {
		t.Fatal(err)
	}

	tf := &DestroyEdgeTransformer{
		Config:  testModule(t, "transform-destroy-edge-module-only"),
		Schemas: simpleTestSchemas(),
	}
	if err := tf.Transform(&g); err != nil {
		t.Fatalf("err: %s", err)
	}

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(`
module.child.test_object.a (destroy)
  module.child.test_object.b (destroy)
  module.child.test_object.c (destroy)
module.child.test_object.b (destroy)
  module.child.test_object.c (destroy)
module.child.test_object.c (destroy)
`)
	if actual != expected {
		t.Fatalf("wrong result\n\ngot:\n%s\n\nwant:\n%s", actual, expected)
	}
}

func testDestroyNode(addrString string) GraphNodeDestroyer {
	instAddr := mustResourceInstanceAddr(addrString)

	inst := NewNodeAbstractResourceInstance(instAddr)

	return &NodeDestroyResourceInstance{NodeAbstractResourceInstance: inst}
}

const testTransformDestroyEdgeBasicStr = `
test_object.A (destroy)
  test_object.B (destroy)
test_object.B (destroy)
`

const testTransformDestroyEdgeCreatorStr = `
test_object.A
  test_object.A (destroy)
test_object.A (destroy)
  test_object.B (destroy)
test_object.B (destroy)
`

const testTransformDestroyEdgeMultiStr = `
test_object.A (destroy)
  test_object.B (destroy)
  test_object.C (destroy)
test_object.B (destroy)
  test_object.C (destroy)
test_object.C (destroy)
`

const testTransformDestroyEdgeSelfRefStr = `
test_object.A (destroy)
`

const testTransformDestroyEdgeModuleStr = `
module.child.test_object.b (destroy)
  test_object.a (destroy)
test_object.a (destroy)
`
