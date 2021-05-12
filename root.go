package raml

// Root is interface for anything that could become
// RAML root document
type Root interface {
	PostProcess(string, string) error
}
