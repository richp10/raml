package raml

// SecuritySchemeMethod is a description of the following security-related
// request components determined by the scheme:
//   the headers, query parameters, or responses
type SecuritySchemeMethod struct {
	Headers         map[HTTPHeader]Header     `yaml:"headers"`
	QueryParameters map[string]NamedParameter `yaml:"queryParameters"`
	QueryString     map[string]NamedParameter `yaml:"queryString"`
	Responses       map[HTTPCode]Response     `yaml:"responses"`
	// TODO annotation
}

// SecurityScheme defines mechanisms to secure data access, identify
// requests, and determine access level and data visibility.
type SecurityScheme struct {
	Name string
	// TODO: Fill this during the post-processing phase

	// The type attribute MAY be used to convey information about
	// authentication flows and mechanisms to processing applications
	// such as Documentation Generators and Client generators.
	// The security schemes property that MUST be used to specify the API security mechanisms,
	// including the required settings and the authentication methods that the API supports.
	// One API-supported authentication method is allowed.
	// The value MUST be one of the following methods:
	//		OAuth 1.0, OAuth 2.0, Basic Authentication, Digest Authentication, Pass Through, x-<other>
	Type string `yaml:"type"`

	// An alternate, human-friendly name for the security scheme.
	DisplayName string `yaml:"displayName"`

	// Information that MAY be used to describe a security scheme.
	// Its value is a string and MAY be formatted using markdown.
	Description string `yaml:"description"`

	// A description of the following security-related request
	// components determined by the scheme:
	// the headers, query parameters, or responses.
	// As a best practice, even for standard security schemes,
	// API designers SHOULD describe these properties of security schemes.
	// Including the security scheme description completes the API documentation.
	DescribedBy SecuritySchemeMethod `yaml:"describedBy"`

	// The settings attribute MAY be used to provide security scheme-specific information.
	Settings map[string]Any `yaml:"settings"`
}
