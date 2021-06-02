package raml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"path"
	"strings"
)

// A Resource is the conceptual mapping to an entity or set of entities.
type Resource struct {

	// Resources are identified by their relative URI, which MUST begin with
	// a slash (/).
	URI string

	// An alternate, human-friendly name for the resource.
	// If the displayName property is not defined for a resource,
	// documentation tools SHOULD refer to the resource by its property key
	// which acts as the resource name. For example, tools should refer to the relative URI /jobs.
	DisplayName string `yaml:"displayName"`

	// A substantial, human-friendly description of a resource.
	// Its value is a string and MAY be formatted using markdown.
	Description string `yaml:"description"`

	// TODO : annotationName

	// In a REST-ful API, methods are operations that are performed on a
	// resource. A method MUST be one of the HTTP methods defined in the
	// HTTP version 1.1 specification [RFC2616] and its extension,
	// RFC5789 [RFC5789].
	Get     *Method `yaml:"get"`
	Patch   *Method `yaml:"patch"`
	Put     *Method `yaml:"put"`
	Head    *Method `yaml:"head"`
	Post    *Method `yaml:"post"`
	Delete  *Method `yaml:"delete"`
	Options *Method `yaml:"options"`

	// A list of traits to apply to all methods declared (implicitly or explicitly) for this resource.
	// Individual methods can override this declaration.
	Is []DefinitionChoice `yaml:"is"`

	// The resource type that this resource inherits.
	Type *DefinitionChoice `yaml:"type"`

	// The security schemes that apply to all methods declared (implicitly or explicitly) for this resource.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`

	// Detailed information about any URI parameters of this resource.
	URIParameters map[string]NamedParameter `yaml:"uriParameters"`

	// A nested resource, which is identified as any property
	// whose name begins with a slash ("/"), and is therefore treated as a relative URI.
	Nested map[string]*Resource `yaml:"-"`

	// A resource defined as a child property of another resource is called a
	// nested resource, and its property's key is its URI relative to its
	// parent resource's URI. If this is not nil, then this resource is a
	// child resource.
	Parent *Resource

	// all methods of this resource
	Methods []*Method `yaml:"-"`
}

func (r *Resource) UnmarshalYAML(node *yaml.Node) error {
	type clone Resource

	c := clone{}
	if err := node.Decode(&c); err != nil {
		return err
	}
	*r = Resource(c)

	var nested = map[string]*Resource{}
	for i, childNode := range node.Content {
		if resourceRegexp.MatchString(childNode.Value) {
			cc := clone{}
			//We fetch the next node, which contains the actual data for the resource
			err := node.Content[i+1].Decode(&cc)
			if err != nil {
				return err
			}
			cc.Parent = r
			nr := Resource(cc)
			nested[childNode.Value] = &nr
		}
	}

	r.Nested = nested

	return nil
}

// postProcess doing post processing of a resource after being constructed by the parser.
// - assign all properties that can't be obtained from RAML document
// - inherit from resource type
// - inherit from traits
func (r *Resource) postProcess(uri string, parent *Resource, resourceTypes map[string]ResourceType,
	traitsMap map[string]Trait, apiDef *APIDefinition) error {
	r.URI = strings.TrimSpace(uri)
	r.Parent = parent

	err := r.setMethods(traitsMap, apiDef)
	if err != nil {
		return err
	}

	// inherit from resource types
	if err = r.inheritResourceType(resourceTypes, apiDef); err != nil {
		return err
	}

	// process nested/child resources
	for k := range r.Nested {
		n := r.Nested[k]
		if err = n.postProcess(k, r, resourceTypes, traitsMap, apiDef); err != nil {
			return err
		}
		r.Nested[k] = n
	}
	return nil
}

// inherit from a resource type
func (r *Resource) inheritResourceType(resourceTypes map[string]ResourceType, apiDef *APIDefinition) error {
	// get resource type object to inherit
	rt, err := r.getResourceType(resourceTypes)
	if rt == nil || err != nil {
		return err
	}

	// initialize dicts
	dicts := initResourceTypeDicts(r, r.Type.Parameters)

	r.Description = substituteParams(r.Description, rt.Description, dicts)

	// uri parameters
	if len(r.URIParameters) == 0 {
		r.URIParameters = map[string]NamedParameter{}
	}
	for name, up := range rt.URIParameters {
		p, ok := r.URIParameters[name]
		if !ok {
			p = NamedParameter{}
		}
		p.inherit(up, dicts)
		r.URIParameters[name] = p
	}

	// methods
	r.inheritMethods(rt, apiDef)

	return nil
}

// inherit methods inherits all methods based on it's resource type
func (r *Resource) inheritMethods(rt *ResourceType, apiDef *APIDefinition) {
	// inherit all methods from resource type
	// if it doesn't have the methods, we create it
	for _, rtm := range rt.methods {
		m := r.MethodByName(rtm.Name)
		if m == nil {
			m = newMethod(rtm.Name)
			r.assignMethod(m, m.Name)
		}
		m.resourceTypeName = r.Type.Name
		m.inheritFromResourceType(r, rtm, apiDef)
	}

	// inherit optional methods if only the resource also has the method
	for _, rtm := range rt.optionalMethods {
		m := r.MethodByName(rtm.Name)
		if m == nil {
			continue
		}
		m.resourceTypeName = r.Type.Name
		m.inheritFromResourceType(r, rtm, apiDef)
	}

}

// get resource type from which this resource will inherit
func (r *Resource) getResourceType(resourceTypes map[string]ResourceType) (*ResourceType, error) {
	// check if it's specify a resource type to inherit
	if r.Type == nil || r.Type.Name == "" {
		return nil, nil
	}

	// get resource type from array of resource type map
	for k, rt := range resourceTypes {
		if k == r.Type.Name {
			return &rt, nil
		}
	}
	return nil, fmt.Errorf("can't find resource type named :%v", r.Type.Name)
}

// set methods set all methods name
// and add it to Methods slice
func (r *Resource) setMethods(traitsMap map[string]Trait, apiDef *APIDefinition) (err error) {
	if r.Get != nil {
		err = r.Get.postProcess(r, "GET", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Post != nil {
		err = r.Post.postProcess(r, "POST", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Put != nil {
		err = r.Put.postProcess(r, "PUT", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Patch != nil {
		err = r.Patch.postProcess(r, "PATCH", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Head != nil {
		err = r.Head.postProcess(r, "HEAD", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Delete != nil {
		err = r.Delete.postProcess(r, "DELETE", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	if r.Options != nil {
		err = r.Options.postProcess(r, "OPTIONS", traitsMap, apiDef)
		if err != nil {
			return err
		}
	}
	return nil
}

// MethodByName return resource's method by it's name
func (r *Resource) MethodByName(name string) *Method {
	switch name {
	case "GET":
		return r.Get
	case "POST":
		return r.Post
	case "PUT":
		return r.Put
	case "PATCH":
		return r.Patch
	case "HEAD":
		return r.Head
	case "DELETE":
		return r.Delete
	case "OPTIONS":
		return r.Options
	default:
		return nil
	}
}

func (r *Resource) assignMethod(m *Method, name string) {
	switch name {
	case "GET":
		r.Get = m
	case "POST":
		r.Post = m
	case "PUT":
		r.Put = m
	case "PATCH":
		r.Patch = m
	case "HEAD":
		r.Head = m
	case "DELETE":
		r.Delete = m
	case "OPTIONS":
		r.Options = m
	default:
		log.Fatalf("assignMethod fatal error, invalid method name:%v", name)
	}
}

// substituteParams substitute all params inside double chevron to the correct value
// param value will be obtained from dicts map
func substituteParams(toReplace, words string, dicts map[string]interface{}) string {
	// non empty scalar node remain unchanged
	// except it has double chevron bracket
	if toReplace != "" && (strings.Index(toReplace, "<<") < 0 && strings.Index(toReplace, ">>") < 0) {
		return toReplace
	}
	if words == "" {
		return toReplace
	}

	removeParamBracket := func(param string) string {
		param = strings.TrimSpace(param)
		return param[2 : len(param)-2]
	}

	// search params
	params := dcRe.FindAllString(words, -1)

	// substitute the params
	for _, p := range params {
		pVal, ok := getParamValue(removeParamBracket(p), dicts)
		if !ok {
			// only replace if param is found
			continue
		}
		words = strings.Replace(words, p, pVal, -1)
	}
	return words
}

// get value of a resource type param
// return false if not exists
func getParamValue(param string, dicts map[string]interface{}) (string, bool) {
	// split between inflectors and real param
	// real param and each inflector is separated by `|`
	cleanParam, inflectors := func() (string, string) {
		arr := strings.SplitN(param, "|", 2)
		if len(arr) != 2 {
			return param, ""
		}
		return strings.TrimSpace(arr[0]), strings.TrimSpace(arr[1])
	}()

	// get the value
	val, ok := func() (string, bool) {
		// get from type parameters
		val, ok := dicts[cleanParam]
		if !ok {
			return "", false
		}
		return fmt.Sprintf("%v", val), true
	}()
	if !ok {
		return "", false
	}

	// inflect the value if needed
	if inflectors != "" {
		for _, inflector := range strings.Split(inflectors, "|") {
			inflector = strings.TrimSpace(inflector)
			var ok bool
			val, ok = doInflect(val, inflector)
			if !ok {
				log.Fatalf("invalid inflector " + inflector)
			}
		}
	}
	return val, true
}

// CleanURI returns URI without `/`, `\`', `{`, and `}`
func (r *Resource) CleanURI() string {
	s := removeDoubleSlash(r.URI)
	return strings.TrimSpace(removeDoubleChevron(s))
}

// FullURI returns full/absolute URI of this resource
func (r *Resource) FullURI() string {
	return doFullURI(r, "")
}

func doFullURI(r *Resource, completeURI string) string {
	completeURI = path.Join(r.URI, completeURI, "/")
	if r.Parent == nil {
		return completeURI
	}
	return doFullURI(r.Parent, completeURI)
}

// from spec : the rightmost of the non-URI-parameter-containing path fragments.
func (r *Resource) resourcePathName() string {
	// remove trailing slash
	uri := strings.TrimSuffix(r.URI, "/")

	if uri != "" && !strings.HasSuffix(uri, "}") {
		// check if it is non-URI params, which ended by "}"
		elements := strings.Split(uri, "/")
		return elements[len(elements)-1]
	}
	if r.Parent == nil {
		return ""
	}
	return r.Parent.resourcePathName()
}

func removeDoubleSlash(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(s, "/"), "/")
}

func removeDoubleChevron(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(s, "}"), "{")
}
