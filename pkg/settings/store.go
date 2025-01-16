package settings

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/mohae/deepcopy"
	yaml "gopkg.in/yaml.v3"
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
	// TODO: provide a file lock
	names := strings.Split(path, ".")
	fname := names[0] + ".yaml"
	bytes, err := os.ReadFile(filepath.Join(y.Path, fname))
	if errors.Is(err, os.ErrNotExist) {
		return errNotExist(path, err)
	} else if err != nil {
		return err
	}
	var root yaml.Node
	if err := yaml.Unmarshal(bytes, &root); err != nil {
		return err
	}
	rootMapping := root.Content[0]
	node, err := findYamlNode(rootMapping, names[1:])
	if err != nil {
		return err
	}
	err = node.Decode(s)
	if err != nil {
		return err
	}
	return nil
}

func (y *YamlStore) Save(path string, s interface{}) error {
	// TODO: provide a file lock
	names := strings.Split(path, ".")
	fname := names[0] + ".yaml"
	fullPath := filepath.Join(y.Path, fname)
	bytes, err := os.ReadFile(fullPath)
	var newRoot *yaml.Node
	if errors.Is(err, os.ErrNotExist) {
		newRoot, err = createYamlNode(names[1:], s)
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
		var root yaml.Node
		if err := yaml.Unmarshal(bytes, &root); err != nil {
			return err
		}
		rootMapping := root.Content[0]
		var err error
		newRoot, err = alterYamlNode(rootMapping, names[1:], s)
		if err != nil {
			return err
		}
	}
	out, err := yaml.Marshal(newRoot)
	if err != nil {
		return err
	}
	// FIXME: perms
	// TODO: wrap error
	return os.WriteFile(fullPath, out, 0644)
}

func findYamlNode(n *yaml.Node, names []string) (*yaml.Node, error) {
	if len(names) == 0 {
		return n, nil
	}

	if n.Kind != yaml.MappingNode {
		// TODO: a custom error
		return nil, fmt.Errorf("settings: %s: expected a mapping node, got Kind(%d)", names[0], n.Kind)
	}

	for i := 0; i < len(n.Content)-1; i += 2 {
		if n.Content[i].Value == names[0] {
			return findYamlNode(n.Content[i+1], names[1:])
		}
	}
	return nil, ErrNotFound
}

func alterYamlNode(n *yaml.Node, names []string, value interface{}) (*yaml.Node, error) {
	// TODO: move into own func
	if len(names) == 0 {
		newNode := &yaml.Node{}
		err := newNode.Encode(value)
		if err != nil {
			return nil, err
		}
		return newNode, nil
	}

	if n == nil {
		return createYamlNode(names, value)
	}
	if n.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("settings: %s: expected a mapping node, got Kind(%d)", names[0], n.Kind)
	}

	for i := 0; i < len(n.Content)-1; i += 2 {
		if n.Content[i].Value == names[0] {
			newNode, err := alterYamlNode(n.Content[i+1], names[1:], value)
			if err != nil {
				return nil, err
			}

			n.Content[i+1] = newNode
			return n, nil
		}
	}

	newNode, err := createYamlNode(names[1:], value)
	if err != nil {
		return nil, err
	}
	keyNode := &yaml.Node{}
	keyNode.SetString(names[0])
	n.Content = append(n.Content, keyNode, newNode)
	return n, nil
}

func createYamlNode(names []string, value interface{}) (*yaml.Node, error) {
	if len(names) == 0 {
		newNode := &yaml.Node{}
		err := newNode.Encode(value)
		if err != nil {
			return nil, err
		}
		return newNode, nil
	}

	nextNode, err := createYamlNode(names[1:], value)
	if err != nil {
		return nil, err
	}
	keyNode := &yaml.Node{}
	keyNode.SetString(names[0])

	mapping := &yaml.Node{}
	mapping.Kind = yaml.MappingNode
	mapping.Content = []*yaml.Node{keyNode, nextNode}
	return mapping, nil
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
