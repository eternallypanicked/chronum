package parser

import "chronum/parser/types"

type Parser interface{
	Parse(src types.ChronumSource) (*types.ChronumDefinition, error)
	Validate(flow *types.ChronumDefinition) error
}