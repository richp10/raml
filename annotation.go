package raml

import (
	"gopkg.in/yaml.v3"
	"regexp"
)

var annotationNameRegexp = regexp.MustCompile(`^\(.*\)$`)

//Annotations contains a map of referenced annotations and their values
type Annotations struct {
	AnnotationNames map[AnnotationName]interface{}
}

func (a *Annotations) UnmarshalYAML(node *yaml.Node) error {
	var annotations = make(map[AnnotationName]interface{})
	for i := 0; i < len(node.Content); i += 2 {
		var keyNode = node.Content[i]
		var valueNode = node.Content[i+1]

		if annotationNameRegexp.MatchString(keyNode.Value) {
			var values interface{}
			switch valueNode.Kind {
			case yaml.MappingNode:
				err := valueNode.Decode(&values)
				if err != nil {
					return err
				}
				break
			case yaml.ScalarNode:
				values = valueNode.Value
				break
			}

			annotations[AnnotationName(keyNode.Value)] = values
		}
	}

	a.AnnotationNames = annotations

	return nil
}

//AnnotationName contains an annotation reference
type AnnotationName string

// AnnotationType describes the annotation
type AnnotationType struct {
	//TODO: fill this in
}
