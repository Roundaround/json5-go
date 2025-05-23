package token

import (
	"fmt"
)

type Kind int

const (
	UNKNOWN Kind = iota
	ILLEGAL
	EOF

	LEFT_BRACE
	RIGHT_BRACE
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	COLON

	LINE_COMMENT
	BLOCK_COMMENT

	UNQUOTED_STRING
	QUOTED_STRING
	BOOLEAN
	NULL

	DECIMAL_NUMBER
	HEX_NUMBER
	INFINITY
	NAN
)

func (k Kind) String() string {
	switch k {
	case ILLEGAL:
		return "Illegal"
	case EOF:
		return "EOF"
	case LEFT_BRACE:
		return "Left Brace"
	case RIGHT_BRACE:
		return "Right Brace"
	case LEFT_BRACKET:
		return "Left Bracket"
	case RIGHT_BRACKET:
		return "Right Bracket"
	case COMMA:
		return "Comma"
	case COLON:
		return "Colon"
	case LINE_COMMENT:
		return "Line Comment"
	case BLOCK_COMMENT:
		return "Block Comment"
	case UNQUOTED_STRING:
		return "Unquoted String"
	case QUOTED_STRING:
		return "Quoted String"
	case BOOLEAN:
		return "Boolean"
	case NULL:
		return "null"
	case DECIMAL_NUMBER:
		return "Decimal Number"
	case HEX_NUMBER:
		return "Hex Number"
	case INFINITY:
		return "Infinity"
	case NAN:
		return "NaN"
	default:
		return "Unknown"
	}
}

type Token struct {
	Kind    Kind
	Literal string
	Offset  int
}

func (t Token) String() string {
	return fmt.Sprintf("%s %q", t.Kind, t.Literal)
}

var keywords = map[string]Kind{
	"true":     BOOLEAN,
	"false":    BOOLEAN,
	"null":     NULL,
	"Infinity": INFINITY,
	"NaN":      NAN,
}

func LookupKeyword(literal string) (Kind, bool) {
	kind, ok := keywords[literal]
	return kind, ok
}
