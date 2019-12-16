package parser

import (
	"os"
)

const (
	fluxParserTypeEnvVar = "FLUX_PARSER_TYPE"
	parserTypeRust       = "rust"
)

var cachedUseRustParser = os.Getenv(fluxParserTypeEnvVar) == parserTypeRust

func useRustParser() bool {
	return cachedUseRustParser
}
