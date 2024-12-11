package settings

import (
	"errors"
	"reflect"
	"testing"
)

type testSection struct {
	Name string `yaml:"name"`
}

const exisingKey = "existing"
const nonexistentKey = "nonexistent"

var existingSection = &testSection{Name: "existing"}
var nonexistentSection = &testSection{Name: "nonexistent"}
var newSection = &testSection{Name: "new"}

var data = map[string]interface{}{
	exisingKey: existingSection,
}

func makeVirtualStore() Store {
	return &VirtualStore{Data: data}
}

func TestAllStores(t *testing.T) {
	type makeStore func() Store
	type testFunc func(t *testing.T, s Store)

	stores := []struct {
		name string
		make makeStore
	}{
		{"Virtual", makeVirtualStore},
	}

	tests := []struct {
		name string
		test testFunc
	}{
		{"LoadExisting", func(t *testing.T, s Store) {
			var section testSection
			if err := s.Load(exisingKey, &section); err != nil {
				t.Fatalf("Load() err = %v", err)
			}
			if !reflect.DeepEqual(&section, existingSection) {
				t.Errorf("Load() = %v, want %v", section, existingSection)
			}
		}},
		{"LoadNonexistent", func(t *testing.T, s Store) {
			var section testSection
			err := s.Load(nonexistentKey, &section)
			if err == nil {
				t.Fatalf("Load() = %v, want error", section)
			}
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("Load() err = %v, want %v", err, ErrNotFound)
			}
		}},
		{"SaveNew", func(t *testing.T, s Store) {
			if err := s.Save(nonexistentKey, newSection); err != nil {
				t.Fatalf("Save() err = %v", err)
			}
			var section testSection
			if err := s.Load(nonexistentKey, &section); err != nil {
				t.Fatalf("Load() err = %v", err)
			}
			if !reflect.DeepEqual(&section, newSection) {
				t.Errorf("Load() = %v, want %v", section, newSection)
			}
		}},
		{"SaveOverwrite", func(t *testing.T, s Store) {
			if err := s.Save(exisingKey, newSection); err != nil {
				t.Fatalf("Save() err = %v", err)
			}
			var section testSection
			if err := s.Load(exisingKey, &section); err != nil {
				t.Fatalf("Load() err = %v", err)
			}
			if !reflect.DeepEqual(&section, newSection) {
				t.Errorf("Load() = %v, want %v", section, newSection)
			}
		}},
	}

	for _, s := range stores {
		for _, tt := range tests {
			name := s.name + "_" + tt.name
			store := s.make()
			t.Run(name, func(t *testing.T) {
				tt.test(t, store)
			})
		}
	}
}
