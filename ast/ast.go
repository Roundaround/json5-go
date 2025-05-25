package ast

import (
	"strconv"
	"strings"

	"github.com/Roundaround/json5-go/path"
)

type Kind int

const (
	UNKNOWN Kind = iota
	OBJECT
	ARRAY
	STRING
	NUMBER
	BOOLEAN
	NULL
	INFINITY
	NAN
	COMMENT
)

func (k Kind) String() string {
	switch k {
	case OBJECT:
		return "Object"
	case ARRAY:
		return "Array"
	case STRING:
		return "String"
	case NUMBER:
		return "Number"
	case BOOLEAN:
		return "Boolean"
	case NULL:
		return "null"
	case INFINITY:
		return "Infinity"
	case NAN:
		return "NaN"
	case COMMENT:
		return "Comment"
	default:
		return "Unknown"
	}
}

type Node interface {
	Kind() Kind
	Offset() int
	Line() int
	Column() int
	Segment() path.Segment
}

type Position struct {
	offset  int
	line    int
	column  int
	segment path.Segment
}

func (p *Position) Offset() int {
	return p.offset
}

func (p *Position) Line() int {
	return p.line
}

func (p *Position) Column() int {
	return p.column
}

func (p *Position) Segment() path.Segment {
	return p.segment
}

type ObjectNode struct {
	values map[string]Node
	Position
}

func (n *ObjectNode) Kind() Kind {
	return OBJECT
}

func (n *ObjectNode) Len() int {
	return len(n.values)
}

func (n *ObjectNode) Values() map[string]Node {
	return n.values
}

func (n *ObjectNode) Value(key string) (Node, bool) {
	value, ok := n.values[key]
	return value, ok
}

type ArrayNode struct {
	values []Node
}

func (n *ArrayNode) Kind() Kind {
	return ARRAY
}

func (n *ArrayNode) Len() int {
	return len(n.values)
}

func (n *ArrayNode) Values() []Node {
	return n.values
}

func (n *ArrayNode) Value(index int) (Node, bool) {
	if index < 0 || index >= len(n.values) {
		return nil, false
	}
	return n.values[index], true
}

func String(literal string) *StringNode {
	quote := rune(0)
	if strings.HasPrefix(literal, "'") {
		quote = '\''
	} else if strings.HasPrefix(literal, "\"") {
		quote = '"'
	}
	return &StringNode{value: literal, quote: quote}
}

type StringNode struct {
	value string
	quote rune
}

func (n *StringNode) Kind() Kind {
	return STRING
}

func (n *StringNode) Value() string {
	return n.value
}

func (n *StringNode) Quote() rune {
	return n.quote
}

type NumberNode struct {
	raw string
}

func (n *NumberNode) Kind() Kind {
	return NUMBER
}

func (n *NumberNode) Int() (int, error) {
	return strconv.Atoi(n.raw)
}

func (n *NumberNode) Int64() (int64, error) {
	return strconv.ParseInt(n.raw, 10, 64)
}

func (n *NumberNode) Float64() (float64, error) {
	return strconv.ParseFloat(n.raw, 64)
}

func (n *NumberNode) Value() (any, error) {
	if strings.Contains(n.raw, ".") {
		return n.Float64()
	}
	return n.Int64()
}

func (n *NumberNode) String() string {
	return n.raw
}

type BooleanNode struct {
	value bool
}

func (n *BooleanNode) Kind() Kind {
	return BOOLEAN
}

func (n *BooleanNode) Value() bool {
	return n.value
}

type NullNode struct {
}

func (n *NullNode) Kind() Kind {
	return NULL
}

func (n *NullNode) Value() any {
	return nil
}
