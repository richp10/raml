package raml

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Property defines a Type property
//TODO: rework property loading to handle the different types
type Property struct {
	Name        string      `yaml:"-"`
	Type        interface{} `yaml:"type"`
	Required    bool        `yaml:"required"`
	Enum        interface{} `yaml:"enum"`
	Description string      `yaml:"description"`

	// string
	Pattern   *string
	MinLength *int
	MaxLength *int

	// number
	Minimum    *float64
	Maximum    *float64
	MultipleOf *float64
	Format     *string

	// array
	MinItems    *int
	MaxItems    *int
	UniqueItems bool
	Items       Items

	_type *Type // pointer to Type of this Property
}

// ToProperty creates a property from an interface we use `interface{}` as property type to support syntactic sugar &
// shortcut using it directly is DEPRECATED
func ToProperty(name string, p interface{}) Property {
	return toProperty(name, p)
}

func toProperty(name string, p interface{}) Property {
	// convert number(int/float) to float
	toFloat64 := func(number interface{}) float64 {
		switch v := number.(type) {
		case int:
			return float64(v)
		case float64:
			return v
		default:
			return v.(float64)
		}
	}
	// convert from map of interface to property
	mapToProperty := func(val map[interface{}]interface{}) Property {
		var p Property
		p.Required = true
		for k, v := range val {
			switch k {
			case "type":
				if p.Format == nil { // if not nil, we already override it
					p.Type = v.(string)
				}
			case "format":
				p.Format = new(string)
				*p.Format = v.(string)
				p.Type = *p.Format
			case "required":
				p.Required = v.(bool)
			case "enum":
				p.Enum = v
			case "description":
				p.Description = v.(string)
			case "minLength":
				p.MinLength = new(int)
				*p.MinLength = v.(int)
			case "maxLength":
				p.MaxLength = new(int)
				*p.MaxLength = v.(int)
			case "pattern":
				p.Pattern = new(string)
				*p.Pattern = v.(string)
			case "minimum":
				p.Minimum = new(float64)
				*p.Minimum = toFloat64(v)
			case "maximum":
				p.Maximum = new(float64)
				*p.Maximum = toFloat64(v)
			case "multipleOf":
				p.MultipleOf = new(float64)
				*p.MultipleOf = toFloat64(v)
			case "minItems":
				p.MinItems = new(int)
				*p.MinItems = v.(int)
			case "maxItems":
				p.MaxItems = new(int)
				*p.MaxItems = v.(int)
			case "uniqueItems":
				p.UniqueItems = v.(bool)
			case "items":
				p.Items = newItems(v)
			case "properties":
				log.Fatalf("Properties field of '%v' should already be deleted. Seems there are unsupported inline type", name)
			}
		}
		return p
	}

	prop := Property{Required: true}
	switch p.(type) {
	case string:
		prop.Type = p.(string)
	case map[interface{}]interface{}:
		prop = mapToProperty(p.(map[interface{}]interface{}))
	case Property:
		prop = p.(Property)
	}

	if prop.Type == "" { // if has no type, we set it as string
		prop.Type = "string"
	}

	prop.Name = name

	// if has "?" suffix, remove the "?" and set required=false
	if strings.HasSuffix(prop.Name, "?") {
		prop.Required = false
		prop.Name = strings.TrimSuffix(prop.Name, "?")
	}
	return prop

}

// TypeString returns string representation
// of the property's type
func (p Property) TypeString() string {
	switch p.Type.(type) {
	case string:
		return p.Type.(string)
	case Type:
		if p._type == nil {
			panic(fmt.Errorf("property '%v' has no parent type", p.Name))
		}
		return p._type.Name + p.Name
	default:
		return "string"
	}
}

// IsEnum returns true if a property is an enum
func (p Property) IsEnum() bool {
	return p.Enum != nil
}

// IsBidimensionalArray returns true if
// this property is a bidimensional array
func (p Property) IsBidimensionalArray() bool {
	return strings.HasSuffix(p.TypeString(), "[][]")
}

// IsArray returns true if it is an array
func (p Property) IsArray() bool {
	return p.Type == arrayType || strings.HasSuffix(p.TypeString(), "[]")
}

// IsUnion returns true if a property is a union
func (p Property) IsUnion() bool {
	return strings.Index(p.TypeString(), "|") > 0
}

// BidimensionalArrayType returns type of the bidimensional array
func (p Property) BidimensionalArrayType() string {
	return strings.TrimSuffix(p.TypeString(), "[][]")
}

// ArrayType returns the type of the array
func (p Property) ArrayType() string {
	if p.Type == arrayType {
		return p.Items.Type
	}
	return strings.TrimSuffix(p.TypeString(), "[]")
}
