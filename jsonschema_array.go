package raml

import (
	"fmt"
	"strings"
)

func newArraySchema(t *Type, typ, name string) JSONSchema {
	js := JSONSchema{
		Name:   name,
		Schema: schemaVer,
		Type:   "array",
		Items:  newArrayItem(getArrayItemsType(t, typ)),
		T: &Type{
			Type: "array",
		},
	}
	if t == nil {
		return js
	}
	js.MaxItems = t.MaxItems
	js.MinItems = t.MinItems
	js.UniqueItems = t.UniqueItems
	return js
}

func isTypeArray(typ string) bool {
	t := Type{Type: typ}
	return typ == "array" || (t.IsArray() && !t.IsBidimensiArray())
}

type arrayItem struct {
	Type string `json:"type,omitempty"`
	Ref  string `json:"$ref,omitempty"`
}

func newArrayItem(typ string) *arrayItem {
	if _, isScalar := scalarTypes[typ]; isScalar {
		return &arrayItem{
			Type: typ,
		}
	}
	return &arrayItem{
		Ref: typ + fileSuffix,
	}
}

func getArrayItemsType(t *Type, typ string) string {
	if typ == "array" {
		return fmt.Sprint(t.Items)
	}
	return strings.TrimSuffix(typ, "[]")
}
