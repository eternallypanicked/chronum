package yaml

import (
    "gopkg.in/yaml.v3"
    "chronum/parser/types"
    "chronum/parser"
    "fmt"
)

type YamlParser struct{}

func (y *YamlParser) Parse(src types.ChronumSource) (*types.ChronumDefinition, error) {
    var flow types.ChronumDefinition
    if err := yaml.Unmarshal(src.Content, &flow); err != nil {
        return nil, fmt.Errorf("yaml parse error: %w", err)
    }
    if err := y.Validate(&flow); err != nil {
        return nil, err
    }
    return &flow, nil
}

func (y *YamlParser) Validate(flow *types.ChronumDefinition) error {
    if flow.Name == "" {
        return fmt.Errorf("flow must have a name")
    }
    if len(flow.Steps) == 0 {
        return fmt.Errorf("flow must have at least one step")
    }
    return nil
}

func init() {
    parser.Register("yaml", &YamlParser{})
}
