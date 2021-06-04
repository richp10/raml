package raml

// DescribedBy is a description of the following security-related
// request components determined by the scheme:
//   the headers, query parameters, or responses
type DescribedBy struct {
	describedByProps `yaml:",inline"`
	Annotations      Annotations `yaml:",inline"`
}

type describedByProps struct {
	// Optional array of Headers, documenting the possible headers that could be accepted.
	Headers map[HTTPHeader]Header `yaml:"headers"`

	// Query parameters, used by the schema to authorize the request. Mutually exclusive with queryString.
	QueryParameters map[string]NamedParameter `yaml:"queryParameters"`

	// The query string used by the schema to authorize the request. Mutually exclusive with queryParameters.
	QueryString map[string]NamedParameter `yaml:"queryString"`

	// An optional array of responses, representing the possible responses that could be sent.
	Responses map[HTTPCode]Response `yaml:"responses"`
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
	DescribedBy DescribedBy `yaml:"describedBy"`

	// The settings attribute MAY be used to provide security scheme-specific information.
	Settings map[string]Any `yaml:"settings"`
}
