package parser

import (
	"os"
)

const (
	fluxParserTypeEnvVar = "FLUX_PARSER_TYPE"
	parserTypeRust       = "rust"
)

func useRustParser() bool {
	return os.Getenv(fluxParserTypeEnvVar) == parserTypeRust
}
