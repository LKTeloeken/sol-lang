package builtin

import _ "embed"

//go:embed src/solrt.c
var SolRT []byte
