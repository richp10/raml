package raml

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// APIDefinition describes the basic information of an API, such as its
// title and base URI, and describes how to define common schema references.
type APIDefinition struct {
	RAMLVersion string `yaml:"-"`

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
	MediaType string `yaml:"mediaType"`

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

	// TODO : annontation types

	// Declarations of security schemes for use within the API.
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes"`

	// The security schemes that apply to every resource and method in the API.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`

	// Imported external libraries for use within the API.
	Uses map[string]string `yaml:"uses"`

	// The resources of the API, identified as relative URIs that begin with a slash (/).
	// A resource property is one that begins with the slash and is either
	// at the root of the API definition or a child of a resource property. For example, /users and /{groupId}.
	Resources map[string]Resource `yaml:",regexp:/.*"`

	Libraries map[string]*Library `yaml:"-"`

	Filename string
}

// PostProcess doing additional processing
// that couldn't be done by yaml parser such as :
// - inheritance
// - setting some additional values not exist in the .raml
// - allocate map fields
func (apiDef *APIDefinition) PostProcess(workDir, fileName string) error {
	apiDef.Filename = path.Join(workDir, fileName)
	// libraries
	apiDef.Libraries = map[string]*Library{}

	if !isURL(fileName) {
		workDir = filepath.Join(workDir, filepath.Dir(fileName))
	}

	for name, path := range apiDef.Uses {
		lib := &Library{Filename: path}
		if _, err := ParseReadFile(workDir, path, lib); err != nil {
			return fmt.Errorf("apiDef.PostProcess() failed to parse library	name=%v, path=%v\n\terr=%v",
				name, path, err)
		}
		apiDef.Libraries[name] = lib
	}

	// traits
	for name, t := range apiDef.Traits {
		t.postProcess(name)
		apiDef.Traits[name] = t
	}

	// resource types
	for name, rt := range apiDef.ResourceTypes {
		rt.postProcess(name, apiDef.Traits, apiDef)
		apiDef.ResourceTypes[name] = rt
	}

	// types
	for name, t := range apiDef.Types {
		t.postProcess(name, apiDef)
		apiDef.Types[name] = t
	}

	// resources
	for k := range apiDef.Resources {
		r := apiDef.Resources[k]
		rts := apiDef.allResourceTypes(apiDef.ResourceTypes, apiDef.Libraries)
		trts := apiDef.allTraits(apiDef.Traits, apiDef.Libraries)
		if err := r.postProcess(k, nil, rts, trts, apiDef); err != nil {
			return err
		}
		apiDef.Resources[k] = r
	}
	return nil
}

// FindLibFile find lbrary dir and file by it's name
// we also search from included library
func (apiDef *APIDefinition) FindLibFile(name string) (string, string) {
	// search in it's document
	if filename, ok := apiDef.Uses[name]; ok {
		return "", filename
	}

	// search in included libraries
	for _, lib := range apiDef.Libraries {
		if filename, ok := lib.Uses[name]; ok {
			return filepath.Dir(lib.Filename), filename
		}
	}
	return "", ""
}

// GetSecurityScheme gets security scheme by it's name
// it also search in included library
func (apiDef *APIDefinition) GetSecurityScheme(name string) (SecurityScheme, bool) {
	var ss SecurityScheme
	var ok bool

	// split library name by '.'
	// if there is '.', it means we need to look from the library
	splitted := strings.Split(strings.TrimSpace(name), ".")

	switch len(splitted) {
	case 1:
		ss, ok = apiDef.SecuritySchemes[name]
	case 2:
		var l *Library
		l, ok = apiDef.Libraries[splitted[0]]
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
func (apiDef *APIDefinition) allResourceTypes(rts map[string]ResourceType, libraries map[string]*Library) map[string]ResourceType {
	if len(rts) == 0 {
		rts = map[string]ResourceType{}
	}
	for libName, l := range libraries {
		for rtName, rt := range l.ResourceTypes {
			rts[libName+"."+rtName] = rt
		}
		// Recursively processing siblings
		if l.Libraries != nil {
			apiDef.allResourceTypes(rts, l.Libraries)
		}
	}
	return rts
}

// allTraits gets all traits that defined in this api definition.
// traits could be from:
// - the root APIDefinition
// - libraries
func (apiDef *APIDefinition) allTraits(trts map[string]Trait, libraries map[string]*Library) map[string]Trait {
	if len(trts) == 0 {
		trts = map[string]Trait{}
	}
	for libName, l := range libraries {
		for trtName, trt := range l.Traits {
			trts[libName+"."+trtName] = trt
		}
		// Recursively processing siblings
		if l.Libraries != nil {
			apiDef.allTraits(trts, l.Libraries)
		}
	}
	return trts
}

// create new type
func (apiDef *APIDefinition) createType(name string, tip interface{},
	inputProps map[interface{}]interface{}) bool {

	created := false
	// check that there is no type with this name
	if _, exist := apiDef.Types[name]; exist {
		return created
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
		Name:       name,
		Type:       tip,
		Properties: props,
	}
	apiDef.Types[name] = t
	created = true
	return created
}
