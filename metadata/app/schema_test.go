package app

import (
	"testing"
)

func TestGetRegistryRefSuccess(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				Spec: map[string]interface{}{
					"uri": "example.com",
				},
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.GetRegistryRef("simple1")
	if r1 == nil || !ok {
		t.Error("Expected registry to contain 'simple1'")
	}

	uri, ok := r1.Spec["uri"]
	if !ok || uri.(string) != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}
}

func TestGetRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				Spec: map[string]interface{}{
					"uri": "example.com",
				},
				Protocol: "github",
			},
		},
	}

	r1, ok := example1.GetRegistryRef("simple2")
	if r1 != nil || ok {
		t.Error("Expected registry to not contain 'simple2'")
	}
}

func TestAddRegistryRefSuccess(t *testing.T) {
	var example1 = Spec{
		Registries: RegistryRefSpecs{},
	}

	r1, err := example1.AddRegistryRef("simple1", "github", "example.com")
	if r1 == nil || err != nil {
		t.Errorf("Expected registry add to succeed:\n%s", err)
	}

	uri1, ok1 := r1.Spec["uri"]
	if !ok1 || uri1.(string) != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}

	// Test that the `Name` field is added properly if it already exists.
	r2, err := example1.AddRegistryRef("simple1", "github", "example.com")
	if r2 == nil || err != nil {
		t.Errorf("Expected registry add to succeed:\n%s", err)
	}

	uri2, ok2 := r2.Spec["uri"]
	if !ok2 || uri2.(string) != "example.com" || r1.Name != "simple1" || r1.Protocol != "github" {
		t.Errorf("Registry did not add correct values:\n%s", r1)
	}
}

func TestAddRegistryRefFailure(t *testing.T) {
	example1 := Spec{
		Registries: RegistryRefSpecs{
			"simple1": &RegistryRefSpec{
				Spec: map[string]interface{}{
					"uri": "example.com",
				},
				Protocol: "github",
			},
		},
	}

	r1, err := example1.AddRegistryRef("", "github", "example.com")
	if r1 != nil || err != ErrRegistryNameInvalid {
		t.Error("Expected registry to fail to add registry with invalid name")
	}

	r2, err := example1.AddRegistryRef("simple1", "fakeProtocol", "example.com")
	if r2 != nil || err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different protocol")
	}

	r3, err := example1.AddRegistryRef("simple1", "github", "fakeUrl")
	if r3 != nil || err != ErrRegistryExists {
		t.Error("Expected registry to fail to add registry with duplicate name and different uri")
	}
}
