package parser

import "fmt"

var registry = map[string]Parser{}

func Register(name string, p Parser) {
    registry[name] = p
}

func Get(name string) (Parser, error) {
    p, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("parser not found: %s", name)
    }
    return p, nil
}
