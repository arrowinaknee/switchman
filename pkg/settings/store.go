package settings

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/mohae/deepcopy"
)

var ErrNotFound = fmt.Errorf("settings: key not found")

type Store interface {
	Load(path string, s interface{}) error
	Save(path string, s interface{}) error
}

type YamlStore struct {
	Path string // path to directory with yaml files
}

func (y *YamlStore) Load(path string, s interface{}) error {
	panic("unimplemented")
}

func (y *YamlStore) Save(path string, s interface{}) error {
	panic("unimplemented")
}

type VirtualStore struct {
	mut  sync.RWMutex
	Data map[string]interface{}
}

func (v *VirtualStore) Load(path string, s interface{}) error {
	v.mut.RLock()
	defer v.mut.RUnlock()

	if v.Data == nil {
		return fmt.Errorf("settings: virtual data not provided")
	}
	if d, ok := v.Data[path]; ok {
		// TODO: allow loading value to a pointer
		if reflect.TypeOf(s) != reflect.TypeOf(d) {
			return fmt.Errorf("settings: type mismatch: %T != %T", s, d)
		}
		c := deepcopy.Copy(d)
		reflect.ValueOf(s).Elem().Set(reflect.ValueOf(c).Elem())
		return nil
	}
	return errNotExist(path, nil)
}

func (v *VirtualStore) Save(path string, s interface{}) error {
	v.mut.Lock()
	defer v.mut.Unlock()

	if v.Data == nil {
		v.Data = make(map[string]interface{})
	}
	v.Data[path] = s
	return nil
}

// add erronous key to ErrNotFound, can also attach e.g. an os error
func errNotExist(key string, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrNotFound, key, err)
	}
	return fmt.Errorf("%w: %s", ErrNotFound, key)
}
