package raml

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	dcRe = regexp.MustCompile(`\<<(.*?)\>>`) // double chevron regex
)

// ResourceType defines a resource type.
// Resource and method declarations are frequently repetitive. For example, if
// an API requires OAuth authentication, the API definition must include the
// access_token query string parameter (which is defined by the queryParameters
// property) in all the API's resource method declarations.
//
// Moreover, there are many advantages to reusing patterns across multiple
// resources and methods. For example, after defining a collection-type
// resource's characteristics, that definition can be applied to multiple
// resources. This use of patterns encouraging consistency and reduces
// complexity for both servers and clients.
//
// A resource type is a partial resource definition that, like a resource, can
// specify a description and methods and their properties. Resources that use
// a resource type inherit its properties, such as its methods.
type ResourceType struct {

	// TODO: Parameters MUST be indicated in resource type and trait definitions
	// by double angle brackets (double chevrons) enclosing the parameter name;
	// for example, "<<tokenName>>".

	// TODO: Parameter values MAY further be transformed by applying one of
	// the following functions:
	// * The !singularize function MUST act on the value of the parameter
	// by a locale-specific singularization of its original value. The only
	// locale supported by this version of RAML is United States English.
	// * The !pluralize function MUST act on the value of the parameter by a
	// locale-specific pluralization of its original value. The only locale
	// supported by this version of RAML is United States English.

	// Name of the resource type
	Name string

	// The OPTIONAL usage property of a resource type provides instructions
	// on how and when the resource type or trait should be used.
	// Documentation generators MUST convey this property
	// as characteristics of the resource and method, respectively.
	// However, the resources and methods MUST NOT inherit the usage property:
	// neither resources nor methods allow a property named usage.
	Usage string

	// Briefly describes what the resource type
	Description string

	// As in Resource.
	URIParameters map[string]NamedParameter `yaml:"uriParameters"`

	// As in Resource.
	BaseURIParameters map[string]NamedParameter `yaml:"baseUriParameters"`

	// A list of traits to apply to all methods declared (implicitly or explicitly) for this resource type.
	// Individual methods can override this declaration.
	Is []DefinitionChoice `yaml:"is"`

	// In a RESTful API, methods are operations that are performed on a
	// resource. A method MUST be one of the HTTP methods defined in the
	// HTTP version 1.1 specification [RFC2616] and its extension,
	// RFC5789 [RFC5789].
	Get     *Method `yaml:"get"`
	Head    *Method `yaml:"head"`
	Post    *Method `yaml:"post"`
	Put     *Method `yaml:"put"`
	Delete  *Method `yaml:"delete"`
	Patch   *Method `yaml:"patch"`
	Options *Method `yaml:"options"`

	// When defining resource types and traits, it can be useful to capture
	// patterns that manifest several levels below the inheriting resource or
	// method, without requiring the creation of the intermediate levels.
	// For example, a resource type definition may describe a body parameter
	// that will be used if the API defines a post method for that resource,
	// but the processing application should not create the post method itself.
	//
	// This optional structure key indicates that the value of the property
	// should be applied if the property name itself (without the question
	// mark) is already defined (whether explicitly or implicitly) at the
	// corresponding level in that resource or method.
	OptionalURIParameters     map[string]NamedParameter `yaml:"uriParameters?"`
	OptionalBaseURIParameters map[string]NamedParameter `yaml:"baseUriParameters?"`
	OptionalGet               *Method                   `yaml:"get?"`
	OptionalHead              *Method                   `yaml:"head?"`
	OptionalPost              *Method                   `yaml:"post?"`
	OptionalPut               *Method                   `yaml:"put?"`
	OptionalDelete            *Method                   `yaml:"delete?"`
	OptionalPatch             *Method                   `yaml:"patch?"`
	OptionalOptions           *Method                   `yaml:"options?"`

	methods         []*Method // all non-nil methods
	optionalMethods []*Method // all non-nil optional methods
}

// postProcess doing post processing of a resource type after being constructed
// by the .raml parser, some of the works:
// - assign all properties that can't be obtained from RAML document
// - inherit from other resource type
// - apply traits
func (rt *ResourceType) postProcess(name string, traitsMap map[string]Trait, apiDef *APIDefinition) {
	rt.Name = name
	rt.setMethods(traitsMap, apiDef)
	rt.setOptionalMethods()

	// TODO : inherit from other resource type

	// TODO : apply traits
}

// set methods set all methods name
// and add it to methods slice
func (rt *ResourceType) setMethods(traitsMap map[string]Trait, apiDef *APIDefinition) {
	if rt.Get != nil {
		rt.Get.Name = "GET"
		rt.Get.inheritFromTraits(nil, append(rt.Is, rt.Get.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Get)
	}
	if rt.Post != nil {
		rt.Post.Name = "POST"
		rt.Post.inheritFromTraits(nil, append(rt.Is, rt.Post.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Post)
	}
	if rt.Put != nil {
		rt.Put.Name = "PUT"
		rt.Put.inheritFromTraits(nil, append(rt.Is, rt.Put.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Put)
	}
	if rt.Patch != nil {
		rt.Patch.Name = "PATCH"
		rt.Patch.inheritFromTraits(nil, append(rt.Is, rt.Patch.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Patch)
	}
	if rt.Head != nil {
		rt.Head.Name = "HEAD"
		rt.Head.inheritFromTraits(nil, append(rt.Is, rt.Head.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Head)
	}
	if rt.Delete != nil {
		rt.Delete.Name = "DELETE"
		rt.Delete.inheritFromTraits(nil, append(rt.Is, rt.Delete.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Delete)
	}
	if rt.Options != nil {
		rt.Options.Name = "OPTIONS"
		rt.Options.inheritFromTraits(nil, append(rt.Is, rt.Options.Is...), traitsMap, apiDef)
		rt.methods = append(rt.methods, rt.Options)
	}
}

// setOptionalMethods set name of all optional methods
// and add it to optionalMethods slice
func (rt *ResourceType) setOptionalMethods() {
	if rt.OptionalGet != nil {
		rt.OptionalGet.Name = "GET"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalGet)
	}
	if rt.OptionalPost != nil {
		rt.OptionalPost.Name = "POST"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalPost)
	}
	if rt.OptionalPut != nil {
		rt.OptionalPut.Name = "PUT"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalPut)
	}
	if rt.OptionalPatch != nil {
		rt.OptionalPatch.Name = "PATCH"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalPatch)
	}
	if rt.OptionalHead != nil {
		rt.OptionalHead.Name = "HEAD"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalHead)
	}
	if rt.OptionalDelete != nil {
		rt.OptionalDelete.Name = "DELETE"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalDelete)
	}
	if rt.OptionalOptions != nil {
		rt.OptionalOptions.Name = "OPTIONS"
		rt.optionalMethods = append(rt.optionalMethods, rt.OptionalOptions)
	}
}

func initResourceTypeDicts(r *Resource, dicts map[string]interface{}) map[string]interface{} {
	if len(dicts) == 0 {
		dicts = map[string]interface{}{}
	}
	if r != nil {
		dicts["resourcePathName"] = r.resourcePathName()
		dicts["resourcePath"] = r.FullURI()
	}
	return dicts
}

// merge name of type with resource type name.
func mergeTypeName(name, rtName string, apiDef *APIDefinition) string {
	if apiDef == nil {
		log.Warning("passing nil API definition to mergeTypeName")
		return name
	}

	// check if this resource type is in library
	splt := strings.Split(rtName, ".")
	if len(splt) < 2 {
		// if not in library, we need to do nothing
		return name
	}

	// get the library object from API definition root object
	libName := splt[0]
	lib, ok := apiDef.Libraries[libName]
	if !ok {
		return name
	}

	// type not exist in the library
	if _, ok := lib.Types[name]; !ok {
		return name
	}

	return strings.Join([]string{libName, name}, ".")
}
