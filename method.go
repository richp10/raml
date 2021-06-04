package raml

import (
	"fmt"
	"strings"
)

// Method are operations that are performed on a resource
type Method struct {
	methodProps `yaml:",inline"`
	Annotations Annotations `yaml:",inline"`

	Name string

	// name of the resource type this method inherited
	resourceTypeName string
}

type methodProps struct {
	// An alternate, human-friendly method name in the context of the resource.
	// If the displayName property is not defined for a method,
	// documentation tools SHOULD refer to the resource by its property key,
	// which acts as the method name.
	DisplayName string `yaml:"displayName"`

	// A longer, human-friendly description of the method in the context of the resource.
	// Its value is a string and MAY be formatted using markdown.
	Description string `yaml:"description"`

	// Detailed information about any query parameters needed by this method.
	// Mutually exclusive with queryString.
	// The queryParameters property is a map in which the key is the query
	// parameter's name, and the value is itself a map specifying the query
	//  parameter's attributes
	QueryParameters map[string]NamedParameter `yaml:"queryParameters"`

	// Detailed information about any request headers needed by this method.
	Headers map[HTTPHeader]Header `yaml:"headers"`

	// The query string needed by this method.
	// Mutually exclusive with queryParameters.
	QueryString map[string]NamedParameter `yaml:"queryString"`

	// Information about the expected responses to a request.
	// Responses MUST be a map of one or more HTTP status codes, where each
	// status code itself is a map that describes that status code.
	Responses map[HTTPCode]Response `yaml:"responses"`

	// A request body that the method admits.
	Bodies Bodies `yaml:"body"`

	// Explicitly specify the protocol(s) used to invoke a method,
	// thereby overriding the protocols set elsewhere,
	// for example in the baseUri or the root-level protocols property.
	Protocols []string `yaml:"protocols"`

	// A list of the traits to apply to this method.
	Is []DefinitionChoice `yaml:"is"`

	// The security schemes that apply to this method.
	SecuredBy []DefinitionChoice `yaml:"securedBy"`
}

func newMethod(name string) *Method {
	return &Method{
		Name: name,
	}
}

// doing post processing that can't be done by YAML parser
func (m *Method) postProcess(r *Resource, name string, traitsMap map[string]Trait, apiDef *APIDefinition) error {
	m.Name = name
	err := m.inheritFromTraits(r, append(r.Is, m.Is...), traitsMap, apiDef)
	if err != nil {
		return err
	}
	r.Methods = append(r.Methods, m)

	// post process the responses
	responses := make(map[HTTPCode]Response)
	for code, resp := range m.Responses {
		resp.postProcess()
		responses[code] = resp
	}
	m.Responses = responses

	// post process request body
	m.Bodies.postProcess()

	return nil
}

// inherit from resource type
// fields need to be inherited:
// - description
// - response
func (m *Method) inheritFromResourceType(r *Resource, rtm *Method, apiDef *APIDefinition) {
	if rtm == nil {
		return
	}
	dicts := initResourceTypeDicts(r, r.Type.Parameters)

	// inherit description
	m.Description = substituteParams(m.Description, rtm.Description, dicts)

	// inherit display name
	m.DisplayName = substituteParams(m.DisplayName, rtm.DisplayName, dicts)

	// inherit bodies
	m.Bodies.inherit(rtm.Bodies, dicts, m.resourceTypeName, apiDef)

	// inherit headers
	m.inheritHeaders(rtm.Headers, dicts)

	// inherit query params
	m.inheritQueryParams(rtm.QueryParameters, dicts)

	// inherit response
	m.inheritResponses(rtm.Responses, dicts, apiDef)

	// inherit protocols
	m.inheritProtocols(rtm.Protocols)
}

// inherit from all traits, inherited traits are:
// - resource level trait
// - method trait
func (m *Method) inheritFromTraits(r *Resource, is []DefinitionChoice, traitsMap map[string]Trait,
	apiDef *APIDefinition) error {
	for _, tDef := range is {
		// acquire traits object
		t, ok := traitsMap[tDef.Name]
		if !ok {
			return fmt.Errorf("invalid traits name:%v", tDef.Name)
		}

		if err := m.inheritFromATrait(r, &t, tDef.Parameters, apiDef); err != nil {
			return err
		}
	}
	return nil
}

// inherit from a trait
// dicts is map of trait parameters values
func (m *Method) inheritFromATrait(r *Resource, t *Trait, dicts map[string]interface{},
	apiDef *APIDefinition) error {
	dicts = initTraitDicts(r, m, dicts)

	m.Description = substituteParams(m.Description, t.Description, dicts)

	m.Bodies.inherit(t.Bodies, dicts, m.resourceTypeName, apiDef)

	m.inheritHeaders(t.Headers, dicts)

	m.inheritResponses(t.Responses, dicts, apiDef)

	m.inheritQueryParams(t.QueryParameters, dicts)

	m.inheritProtocols(t.Protocols)

	// optional bodies
	// optional headers
	// optional responses
	// optional query parameters
	return nil
}

// inheritHeaders inherit method's headers from parent headers.
// parent headers could be from resource type or a trait
func (m *Method) inheritHeaders(parents map[HTTPHeader]Header, dicts map[string]interface{}) {
	m.Headers = inheritHeaders(m.Headers, parents, dicts)
}

// inheritHeaders inherits headers from parents to childs
func inheritHeaders(childs, parents map[HTTPHeader]Header, dicts map[string]interface{}) map[HTTPHeader]Header {
	if len(childs) == 0 {
		childs = map[HTTPHeader]Header{}
	}

	for name, parent := range parents {
		h, ok := childs[name]
		if !ok {
			if optionalTraitProperty(string(name)) { // don't inherit optional property if not exist
				continue
			}
			h = Header{}
		}
		parent.Name = string(name)
		np := NamedParameter(h)
		np.inherit(NamedParameter(parent), dicts)
		childs[name] = Header(np)
	}
	return childs
}

// inheritQueryParams inherit method's query params from parent query params.
// parent query params could be from resource type or a trait
func (m *Method) inheritQueryParams(parents map[string]NamedParameter, dicts map[string]interface{}) {
	if len(m.QueryParameters) == 0 {
		m.QueryParameters = map[string]NamedParameter{}
	}
	for name, parent := range parents {
		qp, ok := m.QueryParameters[name]
		if !ok {
			if optionalTraitProperty(name) { // don't inherit optional property if not exist
				continue
			}
			qp = NamedParameter{Name: name}
		}
		parent.Name = name // parent name is not initialized by the parser
		qp.inherit(parent, dicts)
		m.QueryParameters[qp.Name] = qp
	}

}

// inheritProtocols inherit method's protocols from parent protocols
// parent protocols could be from resource type or a trait
func (m *Method) inheritProtocols(parent []string) {
	for _, p := range parent {
		m.Protocols = appendStrNotExist(p, m.Protocols)
	}
}

// inheritResponses inherit method's responses from parent responses
// parent responses could be from resource type or a trait
func (m *Method) inheritResponses(parent map[HTTPCode]Response, dicts map[string]interface{},
	apiDef *APIDefinition) {
	if len(m.Responses) == 0 { // allocate if needed
		m.Responses = map[HTTPCode]Response{}
	}
	for code, rParent := range parent {
		resp, ok := m.Responses[code]
		if !ok {
			if optionalTraitProperty(fmt.Sprintf("%v", code)) { // don't inherit optional property if not exist
				continue
			}
			resp = Response{HTTPCode: code}
		}
		resp.inherit(rParent, dicts, m.resourceTypeName, apiDef)
		m.Responses[code] = resp
	}

}

// Response property of a method on a resource describes
// the possible responses to invoking that method on that resource.
// The value of responses is an object that has properties named after
// possible HTTP status codes for that method on that resource.
// The property values describe the corresponding responses.
// Each value is a response declaration.
type Response struct {
	annotations Annotations `yaml:",inline"`

	// HTTP status code of the response
	HTTPCode HTTPCode
	// TODO: Fill this during the post-processing phase

	// A substantial, human-friendly description of a response.
	// Its value is a string and MAY be formatted using markdown.
	Description string

	// An API's methods may support custom header values in responses
	// Detailed information about any response headers returned by this method
	Headers map[HTTPHeader]Header `yaml:"headers"`

	// The body of the response
	Bodies Bodies `yaml:"body"`
}

func (resp *Response) postProcess() {
	resp.Bodies.postProcess()
}

// inherit from parent response
func (resp *Response) inherit(parent Response, dicts map[string]interface{}, rtName string,
	apiDef *APIDefinition) {
	resp.Description = substituteParams(resp.Description, parent.Description, dicts)
	resp.Bodies.inherit(parent.Bodies, dicts, rtName, apiDef)
	resp.Headers = inheritHeaders(resp.Headers, parent.Headers, dicts)
}

// Body is the request/response body
// Some method verbs expect the resource to be sent as a request body.
// For example, to create a resource, the request must include the details of
// the resource to create.
// Resources CAN have alternate representations. For example, an API might
// support both JSON and XML representations.
type Body struct {
	mediaType string `yaml:"mediaType"`
	// TODO: Fill this during the post-processing phase

	// The structure of a request or response body MAY be further specified
	// by the schema property under the appropriate media type.
	// The schema key CANNOT be specified if a body's media type is
	// application/x-www-form-urlencoded or multipart/form-data.
	// All parsers of RAML MUST be able to interpret JSON Schema [JSON_SCHEMA]
	// and XML Schema [XML_SCHEMA].
	// Alternatively, the value of the schema field MAY be the name of a schema
	// specified in the root-level schemas property
	Schema string `yaml:"schema"`

	// Brief description
	Description string `yaml:"description"`

	// Example attribute to generate example invocations
	Example string `yaml:"example"`

	Headers map[HTTPHeader]Header `yaml:"headers"`
}

// Bodies is Container of Body types, necessary because of technical reasons.
//TODO: rework this bit to be more reflective of the actual structure of RAML
type Bodies struct {

	// Instead of using a simple map[HTTPHeader]Body for the body
	// property of the Response and Method, we use the Bodies struct. Why?
	// Because some RAML APIs don't use the MIMEType part, instead relying
	// on the mediaType property in the APIDefinition.
	// So, you might see:
	//
	// responses:
	//   200:
	//     body:
	//       example: "some_example" : "123"
	//
	// and also:
	//
	// responses:
	//   200:
	//     body:
	//       application/json:
	//         example: |
	//           {
	//             "some_example" : "123"
	//           }

	// As in the Body type.
	Schema string `yaml:"schema"`

	// As in the Body type.
	Description string `yaml:"description"`

	// As in the Body type.
	//TODO: fix example unmarshalling: https://github.com/raml-org/raml-spec/blob/master/versions/raml-10/raml-10.md/#defining-examples-in-raml
	Example string `yaml:"-"`

	// Resources CAN have alternate representations. For example, an API
	// might support both JSON and XML representations. This is the map
	// between MIME-type and the body definition related to it.
	//TODO: fix yaml unmarshalling here
	ForMIMEType map[string]Body `yaml:"-"`

	// TODO: For APIs without a priori knowledge of the response types for
	// their responses, "*/*" MAY be used to indicate that responses that do
	// not matching other defined data types MUST be accepted. Processing
	// applications MUST match the most descriptive media type first if
	// "*/*" is used.
	ApplicationJSON *BodiesProperty `yaml:"application/json"`

	// Request/response body type
	Type string `yaml:"type"`
}

// IsEmpty returns true if the body is empty
func (b *Bodies) IsEmpty() bool {
	return b.Type == "" && b.ApplicationJSON == nil
}

// inherit inherits bodies properties from a parent bodies
// parent object could be from trait or response type
func (b *Bodies) inherit(parent Bodies, dicts map[string]interface{}, rtName string, apiDef *APIDefinition) {
	b.Schema = substituteParams(b.Schema, parent.Schema, dicts)
	b.Description = substituteParams(b.Description, parent.Description, dicts)
	b.Example = substituteParams(b.Example, parent.Example, dicts)

	b.Type = mergeTypeName(substituteParams(b.Type, parent.Type, dicts), rtName, apiDef)

	// request body
	if parent.ApplicationJSON != nil {
		if b.ApplicationJSON == nil { // allocate if needed
			b.ApplicationJSON = &BodiesProperty{Properties: map[string]interface{}{}}
		} else if b.ApplicationJSON.Properties == nil {
			b.ApplicationJSON.Properties = map[string]interface{}{}
		}

		b.ApplicationJSON.Type = substituteParams(b.ApplicationJSON.TypeString(), parent.ApplicationJSON.TypeString(), dicts)
		// check if type name is in library
		if typeStr, ok := b.ApplicationJSON.Type.(string); ok {
			b.ApplicationJSON.Type = mergeTypeName(typeStr, rtName, apiDef)
		}

		for k, p := range parent.ApplicationJSON.Properties {
			if _, ok := b.ApplicationJSON.Properties[k]; !ok {

				// handle optional properties as described in
				// https://github.com/raml-org/raml-spec/blob/raml-10/versions/raml-10/raml-10.md#optional-properties
				switch {
				case strings.HasSuffix(k, `\?`): // if ended with `\?` we make it optional property
					k = k[:len(k)-2] + "?"
				case strings.HasSuffix(k, "?"): // if only ended with `?`, we can ignore it
					continue
				}
				k = substituteParams(k, k, dicts)
				prop := toProperty(k, p)
				inheritedType := substituteParams(prop.TypeString(), prop.TypeString(), dicts)
				b.ApplicationJSON.Properties[k] = mergeTypeName(inheritedType, rtName, apiDef)
			}
		}
	}

	// TODO : formimeytype
}

func (b *Bodies) postProcess() {
	if b.ApplicationJSON == nil {
		return
	}

	b.ApplicationJSON.postProcess()
}

// BodiesProperty defines a Body's property
type BodiesProperty struct {
	// we use `interface{}` as property type to support syntactic sugar & shortcut
	Properties map[string]interface{} `yaml:"properties"`

	Type interface{}

	Items interface{}
}

// TypeString returns string representation of the type of the body
func (bp BodiesProperty) TypeString() string {
	return interfaceToString(bp.Type)
}

// GetProperty gets property with given name
// from a bodies
func (bp BodiesProperty) GetProperty(name string) Property {
	p, ok := bp.Properties[name]
	if !ok {
		panic(fmt.Errorf("can't find property name %v", name))
	}
	return toProperty(name, p)
}

// - normalize inline array definition
// - TODO : handle inlined type definition as part of
//	 https://github.com/Jumpscale/go-raml/issues/96
func (bp *BodiesProperty) postProcess() {
	bp.normalizeArray()
}

// change this form
// type: array
// items:
//   type: something
//
// to this form
// type: something[]
func (bp *BodiesProperty) normalizeArray() {
	// `type` and `items` can't be nil
	if bp.Type == nil || bp.Items == nil {
		return
	}

	// make sure `type` value = 'array'
	typeStr, ok := bp.Type.(string)
	if !ok && typeStr != arrayType {
		return
	}

	// check items value
	switch item := bp.Items.(type) {
	case string:
		bp.Type = item + "[]"
		bp.Items = nil
	case map[interface{}]interface{}:
		tip, ok := item["type"].(string)
		if !ok {
			return
		}
		bp.Type = tip + "[]"
		delete(item, "type")
		bp.Items = item
	}
}
