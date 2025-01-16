package settings

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestYamlStore_Load(t *testing.T) {
	type obj struct {
		Name string `yaml:"name"`
	}
	files := map[string]string{
		"config.yaml": `
main:
    first:
        name: First nested
    nested:
        name: Nested value
        unused: blah blah
    another: value
top:
    name: Top Level`,
		"second.yaml": `
name: Second value`,
	}

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr error
	}{
		{"topLevel", "config.top", "Top Level", nil},
		{"nested", "config.main.nested", "Nested value", nil},
		{"wholeFile", "second", "Second value", nil},
		{"nofile", "nonexistent", "", ErrNotFound},
		{"nokey", "config.nonexistent", "", ErrNotFound},
		{"nokeyNested", "config.main.nonexistent", "", ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// files will be removed automatically
			dir := t.TempDir()
			for name, text := range files {
				err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0644)
				if err != nil {
					t.Fatalf("WriteFile() err = %v", err)
				}
			}

			store := &YamlStore{Path: dir}
			var gotObj obj
			err := store.Load(tt.key, &gotObj)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Load() err = %q, want %q", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if gotObj.Name != tt.want {
				t.Errorf("Load() = %v, want %v", gotObj.Name, tt.want)
			}
		})
	}
}

func TestYamlStore_Save(t *testing.T) {
	type obj struct {
		Name string `yaml:"name"`
	}
	fileData := map[string]interface{}{
		"config": map[string]any{
			"main": map[string]any{
				"first": obj{
					Name: "First nested",
				},
				"nested": obj{
					Name: "Nested value",
				},
				"another": "value",
			},
			"top": obj{
				Name: "Top Level",
			},
		},
		"second": obj{
			Name: "Second value",
		},
	}

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr error
	}{
		{"topLevel", "config.top", "New Value", nil},
		{"nested", "config.main.nested", "New Value", nil},
		{"wholeFile", "second", "New Value", nil},
		{"newKey", "config.main.new", "New Value", nil},
		{"newTree", "config.more.sub.deeper.new", "New Value", nil},
		{"newFile", "third", "New Value", nil},
		{"newFileNested", "third.main.new", "New Value", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// files will be removed automatically
			dir := t.TempDir()
			for name, data := range fileData {
				bytes, err := yaml.Marshal(data)
				if err != nil {
					t.Fatalf("Marshal() err = %v", err)
				}
				err = os.WriteFile(filepath.Join(dir, name+".yaml"), bytes, 0644)
				if err != nil {
					t.Fatalf("WriteFile() err = %v", err)
				}
			}

			store := &YamlStore{Path: dir}
			err := store.Save(tt.key, &obj{tt.value})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Save() err = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			var gotObj obj
			err = store.Load(tt.key, &gotObj)
			if err != nil {
				t.Fatalf("Load(%q) err = %v", tt.key, err)
			}
			if gotObj.Name != tt.value {
				t.Errorf("Load(%q) = %q, want %q", tt.key, gotObj.Name, tt.value)
			}

			var checkFiles func([]string, map[string]any)
			checkFiles = func(curKey []string, data map[string]any) {
				for k, v := range data {
					if m, ok := v.(map[string]any); ok {
						checkFiles(append(curKey, k), m)
					} else if wantObj, ok := v.(obj); ok {
						key := strings.Join(append(curKey, k), ".")
						if key == tt.key {
							continue
						}
						var gotObj obj
						err := store.Load(key, &gotObj)
						if err != nil {
							t.Errorf("Load(%q) err = %v", key, err)
							continue
						}
						if gotObj.Name != wantObj.Name {
							t.Errorf("Load(%q) = %q, want %q", key, gotObj.Name, wantObj.Name)
						}
					}
				}
			}

			checkFiles([]string{}, fileData)
		})
	}
}
