package raml

import (
	"encoding/json"
	"fmt"
	"sort"
)

const (
	schemaVer  = "http://json-schema.org/schema#"
	fileSuffix = "_schema.json"
)

// JSONSchema represents a json-schema json file
type JSONSchema struct {
	Schema      string              `json:"$schema"`
	Name        string              `json:"-"`
	Description string              `json:"description,omitempty"`
	Type        string              `json:"type"`
	T           *Type               `json:"-"` // underlying RAML type
	Items       *arrayItem          `json:"items,omitempty"`
	Properties  map[string]property `json:"properties,omitempty"`
	Required    []string            `json:"required,omitempty"`

	// Array properties
	MinItems    int  `json:"minItems,omitempty"`
	MaxItems    int  `json:"maxItems,omitempty"`
	UniqueItems bool `json:"uniqueItems,omitempty"`
}

// NewJSONSchema creates JSON schema from an raml type
func NewJSONSchema(t Type, name string) JSONSchema {
	typ := t.TypeString()
	if typ == "" || t.Type == nil {
		typ = "object"
	}

	return NewJSONSchemaFromProps(&t, t.Properties, typ, name)
}

// NewJSONSchemaFromProps creates json schmema
// from a map of properties
func NewJSONSchemaFromProps(t *Type, properties map[string]interface{}, typ, name string) JSONSchema {
	var required []string

	if isTypeArray(typ) {
		return newArraySchema(t, typ, name)
	}

	props := make(map[string]property, len(properties))
	for name := range properties {
		rp := t.GetProperty(name)
		if !isPropTypeSupported(rp) {
			continue
		}

		p := newProperty(rp)
		props[p.Name] = p
		if p.Required {
			required = append(required, p.Name)
		}
	}

	// we need it to be sorted for testing purpose
	sort.Strings(required)
	js := JSONSchema{
		Schema:     schemaVer,
		Name:       name,
		Type:       typ,
		T:          t,
		Properties: props,
		Required:   required,
	}
	if js.T == nil {
		js.T = &Type{
			Type:       typ,
			Properties: properties,
		}
	}
	return js
}

// Inherit inherits JSON schema from the parents
func (js *JSONSchema) Inherit(parents []JSONSchema) {
	// inherit `properties` and `required`
	// only inherit if the property name not exist in
	// child properties
	oriProps := map[string]interface{}{}
	for name, prop := range js.Properties {
		oriProps[name] = prop
	}

	for _, parent := range parents {
		for name, prop := range parent.Properties {
			if _, exist := oriProps[name]; !exist {
				js.Properties[name] = prop
			}
		}
		for _, name := range parent.Required {
			if _, exist := oriProps[name]; !exist {
				if !js.isRequired(name) {
					js.Required = append(js.Required, name)
				}
			}
		}
	}

}

// Supported returns true if the type is supported
func (js JSONSchema) Supported() bool {
	return js.Type == "object" || js.Type == "array"
}
func (js JSONSchema) String() string {
	// for unsupported type, force the type to `object` type
	if !js.Supported() {
		js.Type = "object"
	}

	b, err := json.MarshalIndent(&js, "", "\t")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// RAMLProperties returns all raml property
// of this JSON schema
func (js *JSONSchema) RAMLProperties() map[string]interface{} {
	js.PostUnmarshal()
	props := map[string]interface{}{}
	for name, prop := range js.Properties {
		props[name] = prop.toRAMLProperty()
	}
	return props
}

// PostUnmarshal must be called after
// json.Unmarshal(byte, &jsonSchema)
func (js *JSONSchema) PostUnmarshal() {
	for name, prop := range js.Properties {
		prop.Required = js.isRequired(name)
		js.Properties[name] = prop
	}
}

func (js *JSONSchema) isRequired(propName string) bool {
	for _, name := range js.Required {
		if name == propName {
			return true
		}
	}
	return false
}

type property struct {
	Name     string      `json:"-"`
	Ref      string      `json:"$ref,omitempty"`
	Type     interface{} `json:"type,omitempty"`
	Required bool        `json:"-"`
	Enum     interface{} `json:"enum,omitempty"`

	// string
	MinLength *int    `json:"minLength,omitempty"`
	MaxLength *int    `json:"maxLength,omitempty"`
	Pattern   *string `json:"pattern,omitempty"`

	// number
	Minimum    *float64 `json:"minimum,omitempty"`
	Maximum    *float64 `json:"maximum,omitempty"`
	MultipleOf *float64 `json:"multipleOf,omitempty"`

	// array
	MinItems    *int       `json:"minItems,omitempty"`
	MaxItems    *int       `json:"maxItems,omitempty"`
	UniqueItems bool       `json:"uniqueItems,omitempty"`
	Items       *arrayItem `json:"items,omitempty"`
}

func newProperty(rp Property) property {
	_, isScalar := scalarTypes[rp.TypeString()]

	// complex type
	if rp.Type != "" && !isScalar && !rp.IsArray() && !rp.IsBidimensiArray() {
		return property{
			Name:     rp.Name,
			Ref:      rp.TypeString() + fileSuffix,
			Required: rp.Required,
		}
	}

	mapTypes := func(t string) string {
		typeMap := map[string]string{
			"string":        "string",
			"number":        "number",
			"integer":       "integer",
			"boolean":       "boolean",
			"date-only":     "string",
			"time-only":     "string",
			"datetime-only": "string",
			"datetime":      "string",

			// from number format
			"int8":   "integer",
			"int16":  "integer",
			"int32":  "integer",
			"int64":  "integer",
			"int":    "integer",
			"long":   "integer",
			"float":  "number",
			"double": "number",

			"object": "object",
		}
		if v, ok := typeMap[t]; ok {
			return v
		}
		return t
	}
	p := property{
		Name:        rp.Name,
		Type:        mapTypes(rp.TypeString()),
		Required:    rp.Required,
		Enum:        rp.Enum,
		MinLength:   rp.MinLength,
		MaxLength:   rp.MaxLength,
		Pattern:     rp.Pattern,
		Minimum:     rp.Minimum,
		Maximum:     rp.Maximum,
		MultipleOf:  rp.MultipleOf,
		MinItems:    rp.MinItems,
		MaxItems:    rp.MaxItems,
		UniqueItems: rp.UniqueItems,
	}

	// array
	if rp.IsArray() && !rp.IsBidimensiArray() {
		p.Type = "array"
		p.Items = newArrayItem(rp.ArrayType())
	}

	if !p.Required {
		p.Type = []string{fmt.Sprint(p.Type), "null"}
	}
	return p
}

func (p *property) toRAMLProperty() Property {
	rp := Property{
		Name:        p.Name,
		Type:        p.Type,
		Required:    p.Required,
		MinLength:   p.MinLength,
		MaxLength:   p.MaxLength,
		Pattern:     p.Pattern,
		Minimum:     p.Minimum,
		Maximum:     p.Maximum,
		MultipleOf:  p.MultipleOf,
		MinItems:    p.MinItems,
		UniqueItems: p.UniqueItems,
	}
	return rp
}

func isPropTypeSupported(p Property) bool {
	return !p.IsBidimensiArray() && !p.IsUnion()
}
