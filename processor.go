package raml

// Processor is interface for anything that could become RAML root document
type Processor interface {
	PostProcess(string, string) error
}
