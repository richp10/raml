package raml

// Items represent an RAML "items"
type Items struct {
	Type   string
	Format string
}

func newItems(i interface{}) Items {
	var it Items
	switch v := i.(type) {
	case string:
		it.Type = v
	case map[interface{}]interface{}:
		it.Type = v["type"].(string)
		if f, ok := v["format"].(string); ok {
			it.Format = f
		}
	default:
		panic("unhandled type for new items")
	}
	return it
}
