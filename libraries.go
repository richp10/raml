package raml

import (
	"fmt"
	"path/filepath"
)

// Library is used to combine any collection of data type declarations,
// resource type declarations, trait declarations, and security scheme declarations
// into modular, externalized, reusable groups.
// While libraries are intended to define common declarations in external documents,
// which are then included where needed, libraries can also be defined inline.
type Library struct {
	Types           map[string]Type           `yaml:"types"`
	ResourceTypes   map[string]ResourceType   `yaml:"resourceTypes"`
	Traits          map[string]Trait          `yaml:"traits"`
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes"`
	Uses            map[string]string         `yaml:"uses"`

	// Describes the content or purpose of a specific library.
	// The value is a string and MAY be formatted using markdown.
	Usage string `yaml:"usage"`

	Libraries map[string]*Library `yaml:"-"`
	Filename  string              `yaml:"-"`
}

// PostProcess doing additional processing
// that couldn't be done by yaml parser such as :
// - inheritance
// - setting some additional values not exist in the .raml
// - allocate map fields
func (l *Library) PostProcess(workDir, fileName string) error {
	// libraries
	if !isURL(fileName) {
		workDir = filepath.Join(workDir, filepath.Dir(fileName))
	}
	l.Libraries = map[string]*Library{}
	for name, path := range l.Uses {
		lib := &Library{Filename: path}
		if _, err := ParseReadFile(workDir, path, lib); err != nil {
			return fmt.Errorf("l.PostProcess() failed to parse library	name=%v, path=%v, err=%v",
				name, path, err)
		}
		l.Libraries[name] = lib

	}

	// traits
	for name, t := range l.Traits {
		t.postProcess(name)
		l.Traits[name] = t
	}

	// resource types
	for name, rt := range l.ResourceTypes {
		rt.postProcess(name, l.Traits, nil)
		l.ResourceTypes[name] = rt
	}
	return nil
}
