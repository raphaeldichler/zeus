package formatter

import (
  "github.com/raphaeldichler/zeus/internal/assert"

	"gopkg.in/yaml.v3"
)

var StringToFormat = map[string]Output{
  "json": NewJSONFormatter(),
  "yaml": NewYAMLFormatter(),
  "pretty": NewPrettyFormatter(),
}

type YAML struct{}

func NewYAMLFormatter() *YAML {
	return &YAML{}
}

func (p *YAML) Marshal(obj any) string {
	yamlBytes, err := yaml.Marshal(obj)
	assert.ErrNil(err)

	return string(yamlBytes)
}
