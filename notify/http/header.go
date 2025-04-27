package http

type Header struct {
	Name  string `yaml:"name" json:"name,omitempty" jsonschema:"title=Header Name,description=The name of the header to send"`
	Value string `yaml:"value" json:"value,omitempty" jsonschema:"title=Header Value,description=The value of the header to send"`
}
