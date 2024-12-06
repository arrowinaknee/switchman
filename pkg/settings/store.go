package settings

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/mohae/deepcopy"
)

var ErrNotFound = fmt.Errorf("settings: section not found")

// TODO: just use interface{}, get path from args
// Section is a group of settings that can be loaded and saved separately from others
type Section interface {
	// Path returns full name of section in store
	// Parts of path are separated by '.', and if the path begins with '.',
	// the section will be stored in the main object
	Path() string
}

type Store interface {
	Load(s Section) error
	Save(s Section) error
}

type YamlStore struct {
	Path string // path to directory with yaml files
	Main string // name of main yaml file
}

func (y *YamlStore) Load(s Section) error {
	panic("unimplemented")
}

func (y *YamlStore) Save(s Section) error {
	panic("unimplemented")
}

type VirtualStore struct {
	mut  sync.RWMutex
	Data map[string]Section
}

func (v *VirtualStore) Load(s Section) error {
	v.mut.RLock()
	defer v.mut.RUnlock()

	key := s.Path()
	if key == "" {
		return fmt.Errorf("settings: empty path")
	}
	if v.Data == nil {
		return fmt.Errorf("settings: virtual data not provided")
	}
	if d, ok := v.Data[key]; ok {
		if reflect.TypeOf(s) != reflect.TypeOf(d) {
			return fmt.Errorf("settings: type mismatch: %T != %T", s, d)
		}
		c := deepcopy.Copy(d)
		reflect.ValueOf(s).Elem().Set(reflect.ValueOf(c).Elem())
		return nil
	}
	return errNotExist(key, nil)
}

func (v *VirtualStore) Save(s Section) error {
	v.mut.Lock()
	defer v.mut.Unlock()

	key := s.Path()
	if key == "" {
		return fmt.Errorf("settings: empty path")
	}
	if v.Data == nil {
		v.Data = make(map[string]Section)
	}
	v.Data[key] = s
	return nil
}

func errNotExist(s string, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrNotFound, s, err)
	}
	return fmt.Errorf("%w: %s", ErrNotFound, s)
}
