package raml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"regexp"
	"strings"
)

var resourceRegexp = regexp.MustCompile("^/.*$")

// APIDefinition describes the basic information of an API, such as its
// title and base URI, and describes how to define common schema references.
type APIDefinition struct {
	definitionProps `yaml:",inline"`
	RAMLVersion     string      `yaml:"-"`
	Annotations     Annotations `yaml:",inline"`
}

type definitionProps struct {
	// A short, plain-text label for the API.
	Title string `yaml:"title" validate:"nonzero"`

	// The version of the API, for example "v1"
	Version string `yaml:"version"`

	// A URI that serves as the base for URIs of all resources.
	// Often used as the base of the URL of each resource containing the location of the API.
	// Can be a template URI.
	// The OPTIONAL baseUri property specifies a URI as an identifier for the API as a whole,
	// and MAY be used the specify the URL at which the API is served (its service endpoint),
	// and which forms the base of the URLs of each of its resources.
	// The baseUri property's value is a string that MUST conform to the URI specification RFC2396 or a Template URI.
	BaseURI string `yaml:"baseUri"`

	// Named parameters used in the baseUri (template).
	BaseURIParameters map[string]NamedParameter `yaml:"baseUriParameters"`

	// The protocols supported by the API.
	// The OPTIONAL protocols property specifies the protocols that an API supports.
	// If the protocols property is not explicitly specified, one or more protocols
	// included in the baseUri property is used;
	// if the protocols property is explicitly specified,
	// the property specification overrides any protocol included in the baseUri property.
	// The protocols property MUST be a non-empty array of strings, of values HTTP and/or HTTPS, and is case-insensitive.
	Protocols []string `yaml:"protocols"`

	// The default media types to use for request and response bodies (payloads),
	// for example "application/json".
	// Specifying the OPTIONAL mediaType property sets the default for return by API
	// requests having a body and for the expected responses. You do not need to specify the media type within every body definition.
	// The value of the mediaType property MUST be a sequence of
	// media type strings or a single media type string.
	// The media type applies to requests having a body,
	// the expected responses, and examples using the same sequence of media type strings.
	// Each value needs to conform to the media type specification in RFC6838.
	MediaType MediaType `yaml:"mediaType"`

	// Additional overall documentation for the API.
	// The API definition can include a variety of documents that serve as a
	// user guides and reference documentation for the API. Such documents can
	// clarify how the API works or provide business context.
	// All the sections are in the order in which the documentation is declared.
	Documentation []Documentation `yaml:"documentation"`

	// An alias for the equivalent "types" property for compatibility with RAML 0.8.
	// Deprecated - API definitions should use the "types" property
	// because a future RAML version might remove the "schemas" alias for that property name.
	// The "types" property supports XML and JSON schemas.
	Schemas []map[string]string

	// Declarations of (data) types for use within the API.
	Types map[string]Type `yaml:"types"`

	// Declarations of traits for use within the API.
	Traits map[string]Trait `yaml:"traits"`

	// Declarations of resource types for use within the API.
	ResourceTypes map[string]ResourceType `yaml:"resourceTypes"`

	// Declarations of annotation types for use by Annotations.
	AnnotationTypes map[string]interface{} `yaml:"annotationTypes"`

	// Declarations of security schemes for use within the API.
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes"`

	// The security schemes that apply to every resource and method in the API.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`

	// Imported external libraries for use within the API.
	Uses map[string]string `yaml:"uses"`

	// The resources of the API, identified as relative URIs that begin with a slash (/).
	// A resource property is one that begins with the slash and is either
	// at the root of the API definition or a child of a resource property. For example, /users and /{groupId}.
	Resources map[string]Resource `yaml:"-"`

	Libraries map[string]*Library `yaml:"-"`

	Filename string `yaml:"-"`
}

//UnmarshalYAML will process most fields through the regular decode functionality, but adds extra logic for resources
func (d *APIDefinition) UnmarshalYAML(node *yaml.Node) error {
	type clone APIDefinition

	c := clone{}
	if err := node.Decode(&c); err != nil {
		return err
	}

	if err := node.Decode(&c.Annotations); err != nil {
		return err
	}

	c.definitionProps.Resources = make(map[string]Resource)
	for i := 0; i < len(node.Content); i += 2 {
		var keyNode = node.Content[i]
		var valueNode = node.Content[i+1]

		if resourceRegexp.MatchString(keyNode.Value) {
			var resource = Resource{}
			if valueNode.Kind != yaml.MappingNode {
				continue
			}

			err := valueNode.Decode(&resource)
			if err != nil {
				return err
			}
			c.definitionProps.Resources[keyNode.Value] = resource
		}
	}

	*d = APIDefinition(c)

	return nil
}

// Documentation is the additional overall documentation for the API.
type Documentation struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

//MediaType contains the default media types to use for request and response bodies (payloads)
type MediaType []string

// UnmarshalYAML makes sure we support both a single and sequence of default media types
func (m *MediaType) UnmarshalYAML(node *yaml.Node) (err error) {
	switch node.Kind {
	case yaml.ScalarNode:
		*m = append(*m, node.Value)
		break
	case yaml.SequenceNode:
		for _, c := range node.Content {
			*m = append(*m, c.Value)
		}
		break
	default:
		err = fmt.Errorf("unparsable type %s", node.ShortTag())
	}

	return err
}

// PostProcess doing additional processing
// that couldn't be done by yaml parser such as :
// - inheritance
// - setting some additional values not exist in the .raml
// - allocate map fields
func (d *APIDefinition) PostProcess(workDir, fileName string) error {
	d.Filename = strings.Join([]string{workDir, fileName}, "")
	d.Libraries = map[string]*Library{}

	for name, useFileName := range d.Uses {
		lib := &Library{Filename: strings.Join([]string{workDir, useFileName}, "")}

		if _, err := ParseReadFile(workDir, useFileName, lib); err != nil {
			return fmt.Errorf("d.PostProcess() failed to parse library	name=%v, path=%v\n\terr=%v", name, useFileName, err)
		}
		d.Libraries[name] = lib
	}

	// traits
	for name, t := range d.Traits {
		t.postProcess(name)
		d.Traits[name] = t
	}

	// resource types
	for name, rt := range d.ResourceTypes {
		err := rt.postProcess(name, d.Traits, d)
		if err != nil {
			return err
		}
		d.ResourceTypes[name] = rt
	}

	// types
	for name, t := range d.Types {
		err := t.postProcess(name, d)
		if err != nil {
			return err
		}
		d.Types[name] = t
	}

	// resources
	for k := range d.Resources {
		r := d.Resources[k]
		rts := d.allResourceTypes(d.ResourceTypes, d.Libraries)
		traits := d.allTraits(d.Traits, d.Libraries)
		if err := r.postProcess(k, nil, rts, traits, d); err != nil {
			return err
		}
		d.Resources[k] = r
	}
	return nil
}

// FindLibFile find library dir and file by it's name we also search from included library
func (d *APIDefinition) FindLibFile(name string) (string, string) {
	// search in it's document
	if filename, ok := d.Uses[name]; ok {
		return "", filename
	}

	// search in included libraries
	for _, lib := range d.Libraries {
		if filename, ok := lib.Uses[name]; ok {
			return filepath.Dir(lib.Filename), filename
		}
	}
	return "", ""
}

// GetSecurityScheme gets security scheme by it's name
// it also search in included library
func (d *APIDefinition) GetSecurityScheme(name string) (SecurityScheme, bool) {
	var ss SecurityScheme
	var ok bool

	// split library name by '.'
	// if there is '.', it means we need to look from the library
	splitted := strings.Split(strings.TrimSpace(name), ".")

	switch len(splitted) {
	case 1:
		ss, ok = d.SecuritySchemes[name]
	case 2:
		var l *Library
		l, ok = d.Libraries[splitted[0]]
		if !ok {
			return ss, false
		}
		ss, ok = l.SecuritySchemes[splitted[1]]
	}
	return ss, ok
}

// AllResourceTypes gets all resource type that defined in this api definition.
// resource types could be from:
// - this document itself
// - libraries
func (d *APIDefinition) allResourceTypes(rts map[string]ResourceType, libraries map[string]*Library) map[string]ResourceType {
	if len(rts) == 0 {
		rts = map[string]ResourceType{}
	}
	for libName, l := range libraries {
		for rtName, rt := range l.ResourceTypes {
			rts[libName+"."+rtName] = rt
		}
		// Recursively processing siblings
		if l.Libraries != nil {
			d.allResourceTypes(rts, l.Libraries)
		}
	}
	return rts
}

// allTraits gets all traits that defined in this api definition.
// traits could be from:
// - the root APIDefinition
// - libraries
func (d *APIDefinition) allTraits(traits map[string]Trait, libraries map[string]*Library) map[string]Trait {
	if len(traits) == 0 {
		traits = map[string]Trait{}
	}
	for libName, l := range libraries {
		for trtName, trt := range l.Traits {
			traits[libName+"."+trtName] = trt
		}
		// Recursively processing siblings
		if l.Libraries != nil {
			d.allTraits(traits, l.Libraries)
		}
	}
	return traits
}

// create new type
func (d *APIDefinition) createType(name string, tip interface{},
	inputProps map[interface{}]interface{}) bool {

	// check that there is no type with this name
	if _, exist := d.Types[name]; exist {
		return false
	}

	// convert the inputProps to properties
	props := make(map[string]interface{})

	for k, p := range inputProps {
		name, ok := k.(string)
		if !ok {
			panic(fmt.Errorf("property key:%v need to be a string", k))
		}
		props[name] = p
	}

	t := Type{
		typeProps: typeProps{
			Name:       name,
			Type:       tip,
			Properties: props,
		},
	}
	d.Types[name] = t
	return true
}
