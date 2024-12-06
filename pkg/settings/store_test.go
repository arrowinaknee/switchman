package settings

import "testing"

func makeVirtualStore() Store {
	return &VirtualStore{}
}

func TestAllStores(t *testing.T) {
	type makeStore func() Store
	type testStore func(t *testing.T, s Store)

	stores := []struct {
		name string
		make makeStore
	}{
		{"Virtual", makeVirtualStore},
	}

	tests := []struct {
		name string
		test testStore
	}{}

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

type stubSection struct {
	path string
}

func (s *stubSection) Path() string {
	return s.path
}
